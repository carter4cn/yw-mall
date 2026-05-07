package logic

import (
	"context"
	"database/sql"

	"mall-product-rpc/internal/model"
	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
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

	return &product.CreateProductResp{Id: id}, nil
}
