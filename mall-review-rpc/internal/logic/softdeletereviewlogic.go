package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type SoftDeleteReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSoftDeleteReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SoftDeleteReviewLogic {
	return &SoftDeleteReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SoftDeleteReviewLogic) SoftDeleteReview(in *review.SoftDeleteReviewReq) (*review.Empty, error) {
	// todo: add your logic here and delete this line

	return &review.Empty{}, nil
}
