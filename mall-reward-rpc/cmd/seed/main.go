// Seed binary for mall-reward-rpc.
//
// Registers the reward templates that mall-activity-rpc resolves at runtime
// (codes are stable; ids are not — InnoDB multi-master may skip values, so
// activity-rpc looks up by code via ListTemplates and caches in memory).
//
// Idempotent: CreateTemplate is upsert-on-code so re-running is safe.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-reward-rpc/reward"
)

var rewardAddr = flag.String("reward", "127.0.0.1:9013", "reward rpc address")

type tpl struct {
	code, typ, schema, description string
	maxValue                       int64
}

var templates = []tpl{
	{
		code:        "signin_points",
		typ:         "points",
		description: "每日签到积分奖励",
		schema:      `{"points":{"type":"int","default":10}}`,
		maxValue:    100,
	},
	{
		code:        "lottery_grand_prize",
		typ:         "points",
		description: "抽奖大奖（积分等价物）",
		schema:      `{"points":{"type":"int"}}`,
		maxValue:    10000,
	},
	{
		code:        "coupon_default",
		typ:         "coupon",
		description: "默认优惠券（满 100 减 20）",
		schema:      `{"coupon_code":{"type":"string"},"discount":{"type":"int"}}`,
		maxValue:    100,
	},
	{
		code:        "seckill_order",
		typ:         "physical",
		description: "秒杀实物（DTM SAGA 锁库存 + 创建预订单）",
		schema:      `{"sku_id":{"type":"int64"},"quantity":{"type":"int"}}`,
		maxValue:    1,
	},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(*rewardAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("reward dial: %v", err)
	}
	defer conn.Close()
	cli := reward.NewRewardClient(conn)

	for _, t := range templates {
		res, err := cli.CreateTemplate(ctx, &reward.CreateTemplateReq{
			Code:              t.code,
			Type:              t.typ,
			PayloadSchemaJson: t.schema,
			MaxValue:          t.maxValue,
			Description:       t.description,
		})
		if err != nil {
			fmt.Printf("[reward] CreateTemplate %s: %v\n", t.code, err)
			continue
		}
		fmt.Printf("[reward] template id=%d code=%s type=%s\n", res.Id, t.code, t.typ)
	}

	fmt.Println("\nseed done")
}
