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

type ProductListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductListLogic {
	return &ProductListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductListLogic) ProductList(req *types.ProductListReq) (resp *types.ProductListResp, err error) {
	res, err := l.svcCtx.ProductRpc.ListProducts(l.ctx, &product.ListProductsReq{
		CategoryId: req.CategoryId,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	products := make([]types.ProductDetailResp, 0, len(res.Products))
	for _, p := range res.Products {
		products = append(products, types.ProductDetailResp{
			Id:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CategoryId:  p.CategoryId,
			Images:      p.Images,
			Status:      p.Status,
			CreateTime:  p.CreateTime,
		})
	}

	return &types.ProductListResp{
		Products: products,
		Total:    res.Total,
	}, nil
}
