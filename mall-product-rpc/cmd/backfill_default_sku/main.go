// 一次性脚本：每条 product 自动 INSERT 一行 default SKU
// 复制 product 的 price/stock/images。幂等：UNIQUE(product_id, sku_code) 保护。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := flag.String("dsn",
		"proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_product?charset=utf8mb4&parseTime=true&loc=Local",
		"mall_product DSN")
	flag.Parse()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, shop_id, price, stock, images FROM product")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	inserted, skipped := 0, 0
	for rows.Next() {
		var pid, shopId, price, stock uint64
		var imgs string
		if err := rows.Scan(&pid, &shopId, &price, &stock, &imgs); err != nil {
			log.Fatal(err)
		}
		// product.images 是逗号分隔；取第一张做 SKU 主图
		mainImg := imgs
		if idx := strings.IndexByte(imgs, ','); idx >= 0 {
			mainImg = imgs[:idx]
		}
		code := fmt.Sprintf("P%d-S1", pid)
		res, err := db.Exec(`
            INSERT IGNORE INTO sku
              (product_id, shop_id, sku_code, spec_text, spec_json, price, stock, image, status)
            VALUES (?, ?, ?, '', '', ?, ?, ?, 1)`,
			pid, shopId, code, price, stock, mainImg)
		if err != nil {
			log.Fatalf("insert pid=%d: %v", pid, err)
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			inserted++
		} else {
			skipped++
		}
	}
	fmt.Printf("backfill default sku done: inserted=%d skipped=%d\n", inserted, skipped)
}
