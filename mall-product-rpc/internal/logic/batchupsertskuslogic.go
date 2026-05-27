package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type BatchUpsertSkusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertSkusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertSkusLogic {
	return &BatchUpsertSkusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// BatchUpsertSkus 给商家编辑商品时改 SKU 矩阵用。
// 校验 product 归属 shopId（防止跨店越权改 SKU）。
// 全量上送语义：未在 skus 数组里的现存 SKU 不动；要软删需显式 status=0 → 2。
// 单事务保证 SKU upsert + product 聚合字段同步原子性。
func (l *BatchUpsertSkusLogic) BatchUpsertSkus(in *product.BatchUpsertSkusReq) (*product.BatchUpsertSkusResp, error) {
	if in.ProductId <= 0 || in.ShopId <= 0 {
		return nil, errors.New("product_id and shop_id required")
	}
	var ownerShopId int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &ownerShopId,
		"SELECT shop_id FROM product WHERE id=? LIMIT 1", in.ProductId); err != nil {
		return nil, errors.New("product not found")
	}
	if ownerShopId != in.ShopId {
		return nil, errors.New("product does not belong to this shop")
	}

	var out []*product.SkuItem
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
		items, e := upsertSkusAtomic(ctx, tx, in.ShopId, in.ProductId, in.Skus)
		if e != nil {
			return e
		}
		out = items
		return syncProductAggregateInTx(ctx, tx, in.ProductId)
	})
	if err != nil {
		return nil, err
	}
	return &product.BatchUpsertSkusResp{Items: out}, nil
}
