// Seed binary for mall-activity-rpc:
// 1. Pre-creates 4 sample activities (one per type) bound to the 4 workflow
//    definitions seeded by mall-workflow-rpc/cmd/seed.
// 2. For seckill, inserts an activity_inventory_snapshot row.
// 3. For coupon, primes the Redis stock counter.
// 4. For lottery, primes the Redis prize-pool list.
// 5. PublishActivity flips status DRAFT→PUBLISHED so Participate accepts.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-activity-rpc/activity"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	activityAddr = flag.String("activity", "127.0.0.1:9010", "activity rpc address")
	dataSource   = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_activity?charset=utf8mb4&parseTime=true&loc=Local", "MySQL DSN")
	redisAddr    = flag.String("redis", "127.0.0.1:6379", "Redis addr")
)

// def IDs are looked up by code at runtime — auto-increment values can be
// non-monotonic in dev (cache pre-allocations etc.), so a hard-coded map
// would silently bind activities to the wrong workflow.
var defByCode = map[string]int64{}

type sampleActivity struct {
	code         string
	title        string
	typ          string
	defCode      string
	configJson   string
	seckillStock int64
	couponStock  int64
}

var samples = []sampleActivity{
	{code: "signin_2026", title: "每日签到", typ: "signin", defCode: "signin_v1"},
	{code: "lottery_2026", title: "新春幸运转盘", typ: "lottery", defCode: "lottery_v1"},
	{code: "seckill_2026", title: "秒杀 iPhone", typ: "seckill", defCode: "seckill_v1", seckillStock: 100},
	{code: "coupon_2026", title: "满 100 减 20 优惠券", typ: "coupon", defCode: "coupon_v1", configJson: `{"max_per_user":1}`, couponStock: 500},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn := sqlx.NewMysql(*dataSource)
	rds := redis.MustNewRedis(redis.RedisConf{Host: *redisAddr, Type: "node"})

	// resolve workflow definition codes → ids by direct DB read on workflow DB
	wfConn := sqlx.NewMysql("proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_workflow?charset=utf8mb4&parseTime=true&loc=Local")
	wfRows := []struct {
		Id   int64  `db:"id"`
		Code string `db:"code"`
	}{}
	if err := wfConn.QueryRowsCtx(ctx, &wfRows, "SELECT id, code FROM workflow_definition WHERE code IN ('signin_v1','lottery_v1','seckill_v1','coupon_v1')"); err != nil {
		log.Fatalf("load workflow defs: %v", err)
	}
	for _, r := range wfRows {
		defByCode[r.Code] = r.Id
	}
	if len(defByCode) < 4 {
		log.Fatalf("expected 4 workflow definitions, found %d (%v) — run mall-workflow-rpc/cmd/seed first", len(defByCode), defByCode)
	}
	fmt.Printf("[lookup] workflow defs %v\n", defByCode)

	gconn, err := grpc.NewClient(*activityAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("activity dial: %v", err)
	}
	defer gconn.Close()
	cli := activity.NewActivityClient(gconn)

	now := time.Now().Unix()
	endTime := now + 30*86400

	for _, s := range samples {
		res, err := cli.CreateActivity(ctx, &activity.CreateActivityReq{
			Code:                 s.code,
			Title:                s.title,
			Description:          s.title + "（自动播种）",
			Type:                 s.typ,
			StartTime:            now,
			EndTime:              endTime,
			WorkflowDefinitionId: defByCode[s.defCode],
			ConfigJson:           s.configJson,
		})
		if err != nil {
			fmt.Printf("[activity] CreateActivity %s: %v (skip)\n", s.code, err)
			continue
		}
		actId := res.Id
		fmt.Printf("[activity] created id=%d code=%s type=%s\n", actId, s.code, s.typ)

		// per-type priming
		switch s.typ {
		case "seckill":
			if _, err := conn.ExecCtx(ctx,
				"INSERT INTO `activity_inventory_snapshot`(activity_id, sku_id, total_stock, current_stock) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE total_stock=VALUES(total_stock), current_stock=VALUES(current_stock)",
				actId, 1, s.seckillStock, s.seckillStock,
			); err != nil {
				fmt.Printf("  inventory snapshot: %v\n", err)
			}
		case "coupon":
			key := fmt.Sprintf("activity:%d:coupon_stock", actId)
			if err := rds.SetCtx(ctx, key, fmt.Sprintf("%d", s.couponStock)); err != nil {
				fmt.Printf("  prime coupon stock: %v\n", err)
			}
		case "lottery":
			key := fmt.Sprintf("activity:%d:prizes", actId)
			_, _ = rds.DelCtx(ctx, key)
			// prize pool entries are JSON strings; use RPUSH to seed
			var entries []any
			// crude split — in real use: json.Unmarshal then re-marshal each entry.
			entries = []any{
				`{"weight":90,"remaining":1000,"prize_id":1}`,
				`{"weight":9,"remaining":50,"prize_id":2}`,
				`{"weight":1,"remaining":5,"prize_id":3}`,
			}
			_ = entries
			for _, e := range entries {
				_, _ = rds.RpushCtx(ctx, key, fmt.Sprintf("%v", e))
			}
		}

		// publish
		if _, err := cli.PublishActivity(ctx, &activity.IdReq{Id: actId}); err != nil {
			fmt.Printf("  publish: %v\n", err)
		} else {
			fmt.Printf("  published activity id=%d\n", actId)
		}
	}

	fmt.Println("\nseed done")
}
