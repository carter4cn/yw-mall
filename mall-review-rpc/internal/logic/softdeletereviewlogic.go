package logic

import (
	"context"

	"mall-common/errorx"
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
	rev, err := l.svcCtx.ReviewModel.FindOne(l.ctx, in.ReviewId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ReviewNotFound)
	}
	if rev.Status != 0 {
		return &review.Empty{}, nil
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `review` SET status=1 WHERE id=?", in.ReviewId); err != nil {
		return nil, err
	}
	_, _ = l.svcCtx.Redis.Del(cacheKeyProductSummary(rev.ProductId))
	return &review.Empty{}, nil
}
