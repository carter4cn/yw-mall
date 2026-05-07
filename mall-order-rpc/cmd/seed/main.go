// Seed binary for mall-order-rpc: creates sample orders in various states.
// Requires user addresses (user-rpc seed) and products to exist.
// Idempotent: skips if orders with the same order_no already exist.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-order-rpc/order"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	orderAddr = flag.String("order", "127.0.0.1:9003", "order rpc address")
	orderDS   = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_order?charset=utf8mb4&parseTime=true&loc=Local", "order MySQL DSN")
	userDS    = flag.String("user-ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_user?charset=utf8mb4&parseTime=true&loc=Local", "user MySQL DSN")
	productDS = flag.String("product-ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_product?charset=utf8mb4&parseTime=true&loc=Local", "product MySQL DSN")
)

type orderSeed struct {
	userId     int64
	finalStatus int32 // status to set after creation (0=keep pending)
}

var orderSeeds = []orderSeed{
	{userId: 1, finalStatus: 3}, // completed
	{userId: 1, finalStatus: 0}, // pending
	{userId: 2, finalStatus: 2}, // shipped
	{userId: 2, finalStatus: 1}, // paid
	{userId: 3, finalStatus: 0}, // pending
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	userConn := sqlx.NewMysql(*userDS)
	productConn := sqlx.NewMysql(*productDS)
	orderConn := sqlx.NewMysql(*orderDS)

	// load default addresses per user
	type addrRow struct {
		Id     int64 `db:"id"`
		UserId int64 `db:"user_id"`
	}
	var addrRows []addrRow
	if err := userConn.QueryRowsCtx(ctx, &addrRows,
		"SELECT id, user_id FROM user_address WHERE is_default=1 AND user_id IN (1,2,3)"); err != nil {
		log.Fatalf("load addresses: %v", err)
	}
	addrByUser := make(map[int64]int64)
	for _, r := range addrRows {
		addrByUser[r.UserId] = r.Id
	}
	if len(addrByUser) == 0 {
		log.Fatal("no user addresses found — run mall-user-rpc/cmd/seed first")
	}
	fmt.Printf("[lookup] addresses for %d users\n", len(addrByUser))

	// load a few products to use in orders
	type productRow struct {
		Id    int64  `db:"id"`
		Name  string `db:"name"`
		Price int64  `db:"price"`
	}
	var products []productRow
	if err := productConn.QueryRowsCtx(ctx, &products,
		"SELECT id, name, price FROM product WHERE status=1 ORDER BY id LIMIT 10"); err != nil {
		log.Fatalf("load products: %v", err)
	}
	if len(products) == 0 {
		log.Fatal("no products found — run mall-product-rpc/cmd/seed first")
	}
	fmt.Printf("[lookup] %d products loaded\n", len(products))

	// check existing order count
	var existingCount int64
	if err := orderConn.QueryRowCtx(ctx, &existingCount, "SELECT COUNT(*) FROM `order`"); err != nil {
		log.Fatalf("check orders: %v", err)
	}
	if existingCount >= int64(len(orderSeeds)) {
		fmt.Printf("[skip] %d orders already exist\n", existingCount)
		return
	}

	gconn, err := grpc.NewClient(*orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("order dial: %v", err)
	}
	defer gconn.Close()
	client := order.NewOrderClient(gconn)

	created := 0
	for i, s := range orderSeeds {
		addrId, ok := addrByUser[s.userId]
		if !ok {
			log.Printf("[warn] no default address for user %d, skipping", s.userId)
			continue
		}

		// pick 1-2 products per order
		itemCount := (i % 2) + 1
		items := make([]*order.OrderItem, 0, itemCount)
		for j := 0; j < itemCount && j < len(products); j++ {
			p := products[(i+j)%len(products)]
			items = append(items, &order.OrderItem{
				ProductId:   p.Id,
				ProductName: p.Name,
				Price:       p.Price,
				Quantity:    int32(j + 1),
			})
		}

		resp, err := client.CreateOrder(ctx, &order.CreateOrderReq{
			UserId:    s.userId,
			Items:     items,
			AddressId: addrId,
		})
		if err != nil {
			log.Fatalf("create order user=%d: %v", s.userId, err)
		}
		fmt.Printf("[created] order id=%d no=%s user=%d amount=%d\n",
			resp.Id, resp.OrderNo, s.userId, resp.TotalAmount)

		// update to final status if not pending
		if s.finalStatus > 0 {
			if _, err := client.UpdateOrderStatus(ctx, &order.UpdateOrderStatusReq{
				Id:     resp.Id,
				Status: s.finalStatus,
			}); err != nil {
				log.Printf("[warn] update order %d status %d: %v", resp.Id, s.finalStatus, err)
			} else {
				fmt.Printf("[updated] order %d → status %d\n", resp.Id, s.finalStatus)
			}
		}
		created++
	}

	fmt.Printf("order seed done: created=%d\n", created)
}
