package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// statusOrDefault: SkuInput.status=0 → 默认 active(1)
func statusOrDefault(s int32) int32 {
	if s == 0 {
		return 1
	}
	return s
}

// upsertSkusAtomic 在 transaction 内 upsert SKU 数组：
// - id=0 INSERT (sku_code 空则 P{pid}-S{idx} 自动生成)
// - id>0 UPDATE 全字段
// - status=2 软删
// 返回最终行（含 auto-increment id）。
func upsertSkusAtomic(ctx context.Context, tx sqlx.Session, shopId, productId int64, skus []*product.SkuInput) ([]*product.SkuItem, error) {
	out := make([]*product.SkuItem, 0, len(skus))
	for idx, s := range skus {
		code := s.SkuCode
		if code == "" {
			code = fmt.Sprintf("P%d-S%d", productId, idx+1)
		}
		status := statusOrDefault(s.Status)
		if s.Id == 0 {
			res, err := tx.ExecCtx(ctx, `
                INSERT INTO sku
                  (product_id, shop_id, sku_code, spec_text, spec_json, price, stock, image, status)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				productId, shopId, code, s.SpecText, s.SpecJson,
				s.Price, s.Stock, s.Image, status)
			if err != nil {
				return nil, err
			}
			newId, _ := res.LastInsertId()
			out = append(out, &product.SkuItem{
				Id: newId, ProductId: productId, SkuCode: code,
				SpecText: s.SpecText, SpecJson: s.SpecJson,
				Price: s.Price, Stock: s.Stock, Image: s.Image,
				Status: status,
			})
		} else {
			if _, err := tx.ExecCtx(ctx, `
                UPDATE sku SET sku_code=?, spec_text=?, spec_json=?,
                               price=?, stock=?, image=?, status=?
                WHERE id=? AND product_id=? AND shop_id=?`,
				code, s.SpecText, s.SpecJson,
				s.Price, s.Stock, s.Image, status,
				s.Id, productId, shopId); err != nil {
				return nil, err
			}
			out = append(out, &product.SkuItem{
				Id: s.Id, ProductId: productId, SkuCode: code,
				SpecText: s.SpecText, SpecJson: s.SpecJson,
				Price: s.Price, Stock: s.Stock, Image: s.Image,
				Status: status,
			})
		}
	}
	return out, nil
}

// listSkusByProduct 给 detail 接口用，返 status=1 的活跃 SKU。
func listSkusByProduct(ctx context.Context, db sqlx.SqlConn, productId int64) ([]*product.SkuItem, error) {
	var rows []struct {
		Id        uint64 `db:"id"`
		ProductId uint64 `db:"product_id"`
		SkuCode   string `db:"sku_code"`
		SpecText  string `db:"spec_text"`
		SpecJson  string `db:"spec_json"`
		Price     int64  `db:"price"`
		Stock     int64  `db:"stock"`
		Image     string `db:"image"`
		Status    int32  `db:"status"`
	}
	if err := db.QueryRowsCtx(ctx, &rows, `
        SELECT id, product_id, sku_code, spec_text, spec_json, price, stock, image, status
        FROM sku WHERE product_id=? AND status=1 ORDER BY id`, productId); err != nil {
		return nil, err
	}
	out := make([]*product.SkuItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &product.SkuItem{
			Id: int64(r.Id), ProductId: int64(r.ProductId), SkuCode: r.SkuCode,
			SpecText: r.SpecText, SpecJson: r.SpecJson,
			Price: r.Price, Stock: r.Stock, Image: r.Image, Status: r.Status,
		})
	}
	return out, nil
}

// syncProductAggregateInTx 把多 SKU 聚合回写 product 行：
// price = MIN(active sku price), stock = SUM(active sku stock)
// 这是给 C 端列表/搜索路径用的兼容字段（C 端不读 sku 表）。
func syncProductAggregateInTx(ctx context.Context, tx sqlx.Session, productId int64) error {
	var agg struct {
		MinPrice int64 `db:"min_price"`
		SumStock int64 `db:"sum_stock"`
	}
	if err := tx.QueryRowCtx(ctx, &agg, `
        SELECT COALESCE(MIN(price),0) AS min_price, COALESCE(SUM(stock),0) AS sum_stock
        FROM sku WHERE product_id=? AND status=1`, productId); err != nil {
		return err
	}
	_, err := tx.ExecCtx(ctx,
		"UPDATE product SET price=?, stock=? WHERE id=?",
		agg.MinPrice, agg.SumStock, productId)
	return err
}
