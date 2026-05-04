// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"sync"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-product-rpc/product"
	reviewpb "mall-review-rpc/review"

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

func (l *ProductDetailLogic) ProductDetail(req *types.ProductDetailReq) (*types.ProductDetailResp, error) {
	var (
		productResp *product.GetProductResp
		summary     *reviewpb.RatingSummary
		productErr  error
		wg          sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		productResp, productErr = l.svcCtx.ProductRpc.GetProduct(l.ctx, &product.GetProductReq{Id: req.Id})
	}()
	go func() {
		defer wg.Done()
		summary, _ = l.svcCtx.ReviewRpc.GetProductRatingSummary(l.ctx, &reviewpb.GetProductRatingSummaryReq{ProductId: req.Id})
	}()
	wg.Wait()

	if productErr != nil {
		return nil, productErr
	}
	resp := &types.ProductDetailResp{
		Id:          productResp.Id,
		Name:        productResp.Name,
		Description: productResp.Description,
		Price:       productResp.Price,
		Stock:       productResp.Stock,
		CategoryId:  productResp.CategoryId,
		Images:      productResp.Images,
		Status:      productResp.Status,
		CreateTime:  productResp.CreateTime,
	}
	if summary != nil {
		resp.RatingSummary = &types.GetRatingSummaryResp{
			Avg:            summary.Avg,
			Count:          summary.Count,
			Distribution:   summary.Distribution,
			WithMediaCount: summary.WithMediaCount,
		}
	}
	return resp, nil
}
