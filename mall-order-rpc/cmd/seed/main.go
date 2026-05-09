// Seed binary for mall-order-rpc: creates sample orders in various states.
// Requires user addresses (user-rpc seed) and products to exist.
// Idempotent: skips if orders already exist (count >= expected).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
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

// username-keyed seed specs; user IDs are resolved at runtime.
type orderSeedSpec struct {
	username    string
	finalStatus int32 // 0=keep pending, 1=paid, 2=shipped, 3=completed
}

var orderSeedSpecs = []orderSeedSpec{
	{"alice", 3}, // completed
	{"alice", 0}, // pending
	{"bob", 2},   // shipped
	{"bob", 1},   // paid
	{"demo", 0},  // pending
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	userConn := sqlx.NewMysql(*userDS)
	productConn := sqlx.NewMysql(*productDS)
	orderConn := sqlx.NewMysql(*orderDS)

	// resolve usernames → real DB IDs
	usernames := make([]string, 0, len(orderSeedSpecs))
	seen := map[string]bool{}
	for _, s := range orderSeedSpecs {
		if !seen[s.username] {
			usernames = append(usernames, "'"+s.username+"'")
			seen[s.username] = true
		}
	}
	type userRow struct {
		Id       int64  `db:"id"`
		Username string `db:"username"`
	}
	var userRows []userRow
	if err := userConn.QueryRowsCtx(ctx, &userRows,
		"SELECT id, username FROM `user` WHERE username IN ("+strings.Join(usernames, ",")+")",
	); err != nil {
		log.Fatalf("load users: %v", err)
	}
	userIdByName := make(map[string]int64, len(userRows))
	for _, r := range userRows {
		userIdByName[r.Username] = r.Id
	}
	fmt.Printf("[lookup] users: %v\n", userIdByName)
	if len(userIdByName) == 0 {
		log.Fatal("no users found — run mall-user-rpc/cmd/seed first")
	}

	// load default addresses for the resolved user IDs
	realIds := make([]string, 0, len(userIdByName))
	for _, id := range userIdByName {
		realIds = append(realIds, fmt.Sprintf("%d", id))
	}
	type addrRow struct {
		Id     int64 `db:"id"`
		UserId int64 `db:"user_id"`
	}
	var addrRows []addrRow
	if err := userConn.QueryRowsCtx(ctx, &addrRows,
		"SELECT id, user_id FROM user_address WHERE is_default=1 AND user_id IN ("+strings.Join(realIds, ",")+")",
	); err != nil {
		log.Fatalf("load addresses: %v", err)
	}
	addrByUser := make(map[int64]int64)
	for _, r := range addrRows {
		addrByUser[r.UserId] = r.Id
	}
	fmt.Printf("[lookup] addresses for %d users\n", len(addrByUser))

	// load a few products
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

	var existingCount int64
	if err := orderConn.QueryRowCtx(ctx, &existingCount, "SELECT COUNT(*) FROM `order`"); err != nil {
		log.Fatalf("check orders: %v", err)
	}
	if existingCount >= int64(len(orderSeedSpecs)) {
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
	for i, s := range orderSeedSpecs {
		uid, ok := userIdByName[s.username]
		if !ok {
			log.Printf("[warn] user %q not found, skipping", s.username)
			continue
		}
		addrId, ok := addrByUser[uid]
		if !ok {
			log.Printf("[warn] no default address for user %q (id=%d), skipping", s.username, uid)
			continue
		}

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
			UserId:    uid,
			Items:     items,
			AddressId: addrId,
		})
		if err != nil {
			log.Fatalf("create order user=%s: %v", s.username, err)
		}
		fmt.Printf("[created] order id=%d no=%s user=%s(id=%d) amount=%d\n",
			resp.Id, resp.OrderNo, s.username, uid, resp.TotalAmount)

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
