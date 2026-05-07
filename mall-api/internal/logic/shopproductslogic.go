package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-product-rpc/productclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShopProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShopProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShopProductsLogic {
	return &ShopProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShopProductsLogic) ShopProducts(req *types.ShopProductsReq) (*types.ProductListResp, error) {
	res, err := l.svcCtx.ProductRpc.ListShopProducts(l.ctx, &productclient.ListShopProductsReq{
		ShopId:   req.Id,
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
	return &types.ProductListResp{Products: products, Total: res.Total}, nil
}
