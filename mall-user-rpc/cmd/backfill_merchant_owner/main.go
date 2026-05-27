// 一次性脚本：每条 shop 自动 INSERT 一条 role=owner 的 merchant_staff 记录。
// 幂等：UNIQUE(shop_id, user_id) 保护，重跑无副作用。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	userDSN := flag.String("user-dsn",
		"proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_user?charset=utf8mb4&parseTime=true&loc=Local",
		"mall_user DSN")
	shopDSN := flag.String("shop-dsn",
		"proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_shop?charset=utf8mb4&parseTime=true&loc=Local",
		"mall_shop DSN")
	flag.Parse()

	udb, err := sql.Open("mysql", *userDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer udb.Close()
	sdb, err := sql.Open("mysql", *shopDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer sdb.Close()

	rows, err := sdb.Query("SELECT id, owner_user_id FROM shop WHERE owner_user_id > 0")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	inserted, skipped := 0, 0
	now := time.Now().Unix()
	for rows.Next() {
		var sid, oid uint64
		if err := rows.Scan(&sid, &oid); err != nil {
			log.Fatal(err)
		}
		res, err := udb.Exec(`
            INSERT IGNORE INTO merchant_staff
              (shop_id, user_id, role, status, invited_by, joined_at)
            VALUES (?, ?, 'owner', 1, ?, ?)`, sid, oid, oid, now)
		if err != nil {
			log.Fatalf("insert shop_id=%d owner_id=%d: %v", sid, oid, err)
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			inserted++
		} else {
			skipped++
		}
	}
	fmt.Printf("backfill done: inserted=%d skipped=%d\n", inserted, skipped)
}
