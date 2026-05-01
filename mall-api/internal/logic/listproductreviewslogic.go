// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	reviewpb "mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductReviewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductReviewsLogic {
	return &ListProductReviewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductReviewsLogic) ListProductReviews(req *types.ListProductReviewsReq) (*types.ListProductReviewsResp, error) {
	r, err := l.svcCtx.ReviewRpc.ListProductReviews(l.ctx, &reviewpb.ListProductReviewsReq{
		ProductId: req.ProductId,
		Sort:      req.Sort,
		Score:     req.Score,
		WithMedia: req.WithMedia,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.ReviewItem, 0, len(r.Reviews))
	for _, item := range r.Reviews {
		out = append(out, protoReviewToType(item))
	}
	return &types.ListProductReviewsResp{Reviews: out, Total: r.Total}, nil
}
