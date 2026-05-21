// One-shot：扫 user 表，把明文 phone 列还没迁移的行（phone != '' AND phone_enc = ''）
// 转成 phone_enc + phone_hash。幂等 — 已迁移行 phone_enc != '' 排除。
// 不动 email 列（老库没数据）。
package main

import (
	"context"
	"flag"
	"log"

	"mall-common/cryptox"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ds = flag.String("ds",
	"proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_user?charset=utf8mb4&parseTime=true&loc=Local",
	"mall_user MySQL DSN")

func main() {
	flag.Parse()
	cryptox.MustInit()

	db := sqlx.NewMysql(*ds)
	ctx := context.Background()

	type row struct {
		Id    uint64 `db:"id"`
		Phone string `db:"phone"`
	}
	var rows []*row
	if err := db.QueryRowsCtx(ctx, &rows,
		"SELECT id, phone FROM `user` WHERE phone != '' AND phone_enc = ''"); err != nil {
		log.Fatalf("scan: %v", err)
	}
	log.Printf("found %d rows to migrate", len(rows))

	migrated := 0
	for _, r := range rows {
		enc, err := cryptox.Encrypt(r.Phone)
		if err != nil {
			log.Printf("[skip] id=%d encrypt: %v", r.Id, err)
			continue
		}
		hash := cryptox.Hmac(r.Phone)
		if _, err := db.ExecCtx(ctx,
			"UPDATE `user` SET phone_enc=?, phone_hash=? WHERE id=?",
			enc, hash, r.Id); err != nil {
			log.Printf("[skip] id=%d update: %v", r.Id, err)
			continue
		}
		log.Printf("[ok] id=%d phone=%s", r.Id, r.Phone)
		migrated++
	}
	log.Printf("done: total=%d migrated=%d skipped=%d", len(rows), migrated, len(rows)-migrated)
}
