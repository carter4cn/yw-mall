package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserReviewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserReviewsLogic {
	return &ListUserReviewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserReviewsLogic) ListUserReviews(in *review.ListUserReviewsReq) (*review.ListProductReviewsResp, error) {
	// todo: add your logic here and delete this line

	return &review.ListProductReviewsResp{}, nil
}
