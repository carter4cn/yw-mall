// Seed binary for mall-shop-rpc: creates 6 sample shops if they don't exist.
// Idempotent: skips creation when a shop with the same name already exists.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-shop-rpc/shop"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	shopAddr   = flag.String("shop", "127.0.0.1:9017", "shop rpc address")
	dataSource = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_shop?charset=utf8mb4&parseTime=true&loc=Local", "MySQL DSN")
	minioBase  = flag.String("minio", "http://127.0.0.1:9000/mall-shop", "MinIO base URL for shop images")
)

type shopSeed struct {
	name        string
	description string
	rating      float64
}

var shops = []shopSeed{
	{"数码旗舰店", "专注优质数码产品，品类齐全，正品保障", 4.8},
	{"时尚女装馆", "引领潮流时尚，精选优质女装，每日上新", 4.7},
	{"运动户外专营", "专业运动装备，户外探险首选，质量可靠", 4.6},
	{"家居生活馆", "温馨家居用品，让生活更美好，品质保证", 4.9},
	{"美食零食铺", "精选全球美食，健康美味零食，进口直供", 4.5},
	{"图书文具店", "丰富图书资源，优质文具用品，学习好帮手", 4.7},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn := sqlx.NewMysql(*dataSource)

	gconn, err := grpc.NewClient(*shopAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("shop dial: %v", err)
	}
	defer gconn.Close()
	client := shop.NewShopServiceClient(gconn)

	for i, s := range shops {
		// idempotency: skip if a shop with this name already exists
		var count int64
		if err := conn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM shop WHERE name = ?", s.name); err != nil {
			log.Fatalf("check shop %q: %v", s.name, err)
		}
		if count > 0 {
			fmt.Printf("[skip] shop %q already exists\n", s.name)
			continue
		}

		logo := fmt.Sprintf("%s/logo_%d.png", *minioBase, i+1)
		banner := fmt.Sprintf("%s/banner_%d.png", *minioBase, i+1)

		resp, err := client.CreateShop(ctx, &shop.CreateShopReq{
			Name:        s.name,
			Logo:        logo,
			Banner:      banner,
			Description: s.description,
			Rating:      s.rating,
		})
		if err != nil {
			log.Fatalf("create shop %q: %v", s.name, err)
		}
		fmt.Printf("[created] shop %q id=%d logo=%s\n", s.name, resp.Id, logo)
	}

	fmt.Println("shop seed done")
}
