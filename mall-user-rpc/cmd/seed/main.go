// Seed binary for mall-user-rpc:
// 1. Creates alice / bob / demo accounts directly via SQL + bcrypt.
// 2. Seeds sample delivery addresses for each user.
// Uses direct DB access to avoid the 2-second gRPC server timeout
// that bcrypt at DefaultCost exceeds on CPU-constrained containers.
// Idempotent: skips rows that already exist.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	dataSource = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_user?charset=utf8mb4&parseTime=true&loc=Local", "MySQL DSN")
)

type userSeed struct {
	username string
	password string
	phone    string
}

var seedUsers = []userSeed{
	{"alice", "alice123", "13800138001"},
	{"bob", "bob123", "13900139002"},
	{"demo", "demo123", "13700137003"},
}

type addrSeed struct {
	username     string
	receiverName string
	phone        string
	province     string
	city         string
	district     string
	detail       string
	isDefault    int64
}

var addresses = []addrSeed{
	{"alice", "张三", "13800138001", "北京市", "北京市", "朝阳区", "朝阳区建国路88号国贸大厦B座1001", 1},
	{"alice", "张三", "13800138001", "上海市", "上海市", "浦东新区", "浦东新区陆家嘴环路1000号汇亚大厦2201", 0},
	{"bob", "李四", "13900139002", "广东省", "广州市", "天河区", "天河区天河路385号太古汇1期T1-2601", 1},
	{"bob", "李四", "13900139002", "广东省", "深圳市", "南山区", "南山区科技园科发路15号大族广场A座3301", 0},
	{"demo", "王五", "13700137003", "浙江省", "杭州市", "西湖区", "西湖区文三路478号华星时代广场1501", 1},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn := sqlx.NewMysql(*dataSource)

	// Step 1: ensure user accounts exist.
	userIdByName := map[string]int64{}
	for _, u := range seedUsers {
		var id int64
		err := conn.QueryRowCtx(ctx, &id,
			"SELECT id FROM `user` WHERE username=? LIMIT 1", u.username)
		if err == nil {
			fmt.Printf("[existing] user id=%d username=%s\n", id, u.username)
			userIdByName[u.username] = id
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("bcrypt %s: %v", u.username, err)
		}
		result, err := conn.ExecCtx(ctx,
			"INSERT IGNORE INTO `user` (username, password, phone, avatar) VALUES (?,?,?,'')",
			u.username, string(hash), u.phone)
		if err != nil {
			log.Fatalf("insert user %s: %v", u.username, err)
		}
		newId, _ := result.LastInsertId()
		if newId == 0 {
			// INSERT IGNORE skipped a duplicate; re-fetch
			if err2 := conn.QueryRowCtx(ctx, &newId,
				"SELECT id FROM `user` WHERE username=? LIMIT 1", u.username); err2 != nil {
				log.Fatalf("re-fetch user %s: %v", u.username, err2)
			}
		}
		fmt.Printf("[created] user id=%d username=%s\n", newId, u.username)
		userIdByName[u.username] = newId
	}

	// Step 2: seed delivery addresses.
	now := time.Now().Unix()
	created, skipped := 0, 0
	for _, a := range addresses {
		uid, ok := userIdByName[a.username]
		if !ok {
			fmt.Printf("[skip] address for %s — user not resolved\n", a.username)
			skipped++
			continue
		}

		var count int64
		if err := conn.QueryRowCtx(ctx, &count,
			"SELECT COUNT(*) FROM user_address WHERE user_id=? AND detail=?",
			uid, a.detail); err != nil {
			log.Fatalf("check address: %v", err)
		}
		if count > 0 {
			fmt.Printf("[skip] address for user %d %q already exists\n", uid, a.detail)
			skipped++
			continue
		}

		if _, err := conn.ExecCtx(ctx,
			"INSERT INTO user_address (user_id, receiver_name, phone, province, city, district, detail, is_default, create_time, update_time) VALUES (?,?,?,?,?,?,?,?,?,?)",
			uid, a.receiverName, a.phone, a.province, a.city, a.district, a.detail, a.isDefault, now, now,
		); err != nil {
			log.Fatalf("insert address user=%d: %v", uid, err)
		}
		fmt.Printf("[created] address user=%d %s %s\n", uid, a.city, a.district)
		created++
	}

	fmt.Printf("user seed done: created=%d skipped=%d\n", created, skipped)
}
