package logic

import (
	"context"

	"mall-common/errorx"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplyReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplyReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyReviewLogic {
	return &ReplyReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReplyReviewLogic) ReplyReview(in *review.ReplyReviewReq) (*review.Empty, error) {
	if len(in.Text) < 5 || len(in.Text) > 500 {
		return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
	}
	rev, err := l.svcCtx.ReviewModel.FindOne(l.ctx, in.ReviewId)
	if err != nil || rev.Status != 0 {
		return nil, errorx.NewCodeError(errorx.ReviewNotFound)
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `review` SET merchant_reply_text=?, merchant_reply_time=NOW(), merchant_user_id=? WHERE id=? AND status=0",
		in.Text, in.MerchantUserId, in.ReviewId); err != nil {
		return nil, err
	}
	return &review.Empty{}, nil
}
