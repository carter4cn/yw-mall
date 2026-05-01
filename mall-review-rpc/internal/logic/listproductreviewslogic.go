package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductReviewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductReviewsLogic {
	return &ListProductReviewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductReviewsLogic) ListProductReviews(in *review.ListProductReviewsReq) (*review.ListProductReviewsResp, error) {
	// todo: add your logic here and delete this line

	return &review.ListProductReviewsResp{}, nil
}
