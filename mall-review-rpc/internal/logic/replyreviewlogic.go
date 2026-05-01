package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &review.Empty{}, nil
}
