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

type ListUserReviewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserReviewsLogic {
	return &ListUserReviewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserReviewsLogic) ListUserReviews(req *types.ListUserReviewsReq) (*types.ListProductReviewsResp, error) {
	r, err := l.svcCtx.ReviewRpc.ListUserReviews(l.ctx, &reviewpb.ListUserReviewsReq{
		UserId:   currentUserId(l.ctx),
		Page:     req.Page,
		PageSize: req.PageSize,
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
