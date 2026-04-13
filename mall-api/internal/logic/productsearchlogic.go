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

type ProductSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductSearchLogic {
	return &ProductSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductSearchLogic) ProductSearch(req *types.ProductSearchReq) (resp *types.ProductListResp, err error) {
	res, err := l.svcCtx.ProductRpc.SearchProducts(l.ctx, &product.SearchProductsReq{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
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
