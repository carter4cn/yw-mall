// Seed binary for mall-user-rpc: creates sample addresses for test users 1-3.
// Idempotent: skips if an address with the same detail already exists for the user.
// Run after the main user seed that creates user records.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-user-rpc/user"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	userAddr   = flag.String("user", "127.0.0.1:9000", "user rpc address")
	dataSource = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_user?charset=utf8mb4&parseTime=true&loc=Local", "MySQL DSN")
)

type addrSeed struct {
	userId       int64
	receiverName string
	phone        string
	province     string
	city         string
	district     string
	detail       string
	isDefault    bool
}

var addresses = []addrSeed{
	{1, "张三", "13800138001", "北京市", "北京市", "朝阳区", "朝阳区建国路88号国贸大厦B座1001", true},
	{1, "张三", "13800138001", "上海市", "上海市", "浦东新区", "浦东新区陆家嘴环路1000号汇亚大厦2201", false},
	{2, "李四", "13900139002", "广东省", "广州市", "天河区", "天河区天河路385号太古汇1期T1-2601", true},
	{2, "李四", "13900139002", "广东省", "深圳市", "南山区", "南山区科技园科发路15号大族广场A座3301", false},
	{3, "王五", "13700137003", "浙江省", "杭州市", "西湖区", "西湖区文三路478号华星时代广场1501", true},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn := sqlx.NewMysql(*dataSource)

	gconn, err := grpc.NewClient(*userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("user dial: %v", err)
	}
	defer gconn.Close()
	client := user.NewUserClient(gconn)

	created, skipped := 0, 0
	for _, a := range addresses {
		var count int64
		if err := conn.QueryRowCtx(ctx, &count,
			"SELECT COUNT(*) FROM user_address WHERE user_id=? AND detail=?",
			a.userId, a.detail); err != nil {
			log.Fatalf("check address: %v", err)
		}
		if count > 0 {
			fmt.Printf("[skip] address for user %d %q already exists\n", a.userId, a.detail)
			skipped++
			continue
		}

		resp, err := client.AddAddress(ctx, &user.AddAddressReq{
			UserId:       a.userId,
			ReceiverName: a.receiverName,
			Phone:        a.phone,
			Province:     a.province,
			City:         a.city,
			District:     a.district,
			Detail:       a.detail,
			IsDefault:    a.isDefault,
		})
		if err != nil {
			log.Fatalf("add address user=%d: %v", a.userId, err)
		}
		fmt.Printf("[created] address id=%d user=%d %s %s\n", resp.Id, a.userId, a.city, a.district)
		created++
	}

	fmt.Printf("address seed done: created=%d skipped=%d\n", created, skipped)
}
