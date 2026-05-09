package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductLogic) GetProduct(in *product.GetProductReq) (*product.GetProductResp, error) {
	p, err := l.svcCtx.ProductModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	description := ""
	if p.Description.Valid {
		description = p.Description.String
	}

	return &product.GetProductResp{
		Id:          int64(p.Id),
		Name:        p.Name,
		Description: description,
		Price:       p.Price,
		Stock:       p.Stock,
		CategoryId:  int64(p.CategoryId),
		Images:      p.Images,
		ShopId:      int64(p.ShopId),
		Status:      int32(p.Status),
		CreateTime:  p.CreateTime.Unix(),
	}, nil
}
