// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductDetailLogic) ProductDetail(req *types.ProductDetailReq) (resp *types.ProductDetailResp, err error) {
	res, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &product.GetProductReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.ProductDetailResp{
		Id:          res.Id,
		Name:        res.Name,
		Description: res.Description,
		Price:       res.Price,
		Stock:       res.Stock,
		CategoryId:  res.CategoryId,
		Images:      res.Images,
		Status:      res.Status,
		CreateTime:  res.CreateTime,
	}, nil
}
