package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"mall-product-rpc/internal/model"
	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProductLogic) CreateProduct(in *product.CreateProductReq) (*product.CreateProductResp, error) {
	p := &model.Product{
		Name: in.Name,
		Description: sql.NullString{
			String: in.Description,
			Valid:  in.Description != "",
		},
		Price:      in.Price,
		Stock:      in.Stock,
		CategoryId: uint64(in.CategoryId),
		Images:     in.Images,
		ShopId:     uint64(in.ShopId),
		Status:     1,
	}

	result, err := l.svcCtx.ProductModel.Insert(l.ctx, p)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if in.ShopId > 0 {
		if _, e := l.svcCtx.ShopRpc.IncrProductCount(l.ctx, &shopservice.IncrProductCountReq{
			ShopId: in.ShopId,
			Delta:  1,
		}); e != nil {
			l.Logger.Errorf("IncrProductCount shop=%d err=%v", in.ShopId, e)
		}
	}

	// M2: brand/detail 字段不在 generated Product model 里，raw SQL UPDATE 补齐
	if in.Brand != "" || in.Detail != "" {
		if _, e := l.svcCtx.DB.ExecCtx(l.ctx,
			"UPDATE product SET brand=?, detail=? WHERE id=?",
			in.Brand, in.Detail, id); e != nil {
			l.Logger.Errorf("update brand/detail pid=%d err=%v", id, e)
		}
	}

	// M2: SKU 落地。
	// in.Skus 为空 → 建 1 个 default SKU 兼容老路径（C 端列表读 product.price/stock）
	// 非空 → 单事务 upsert + 聚合回写 product
	if len(in.Skus) == 0 {
		mainImg := in.Images
		if comma := strings.IndexByte(in.Images, ','); comma >= 0 {
			mainImg = in.Images[:comma]
		}
		if _, e := l.svcCtx.DB.ExecCtx(l.ctx, `
            INSERT INTO sku (product_id, shop_id, sku_code, spec_text, spec_json,
                             price, stock, image, status)
            VALUES (?, ?, ?, '', '', ?, ?, ?, 1)`,
			id, in.ShopId, fmt.Sprintf("P%d-S1", id),
			in.Price, in.Stock, mainImg); e != nil {
			l.Logger.Errorf("insert default sku pid=%d err=%v", id, e)
		}
	} else {
		if e := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
			if _, ue := upsertSkusAtomic(ctx, tx, in.ShopId, id, in.Skus); ue != nil {
				return ue
			}
			return syncProductAggregateInTx(ctx, tx, id)
		}); e != nil {
			l.Logger.Errorf("upsert skus pid=%d err=%v", id, e)
			return nil, e
		}
	}

	return &product.CreateProductResp{Id: id}, nil
}
