package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitFollowupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitFollowupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitFollowupLogic {
	return &SubmitFollowupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitFollowupLogic) SubmitFollowup(in *review.SubmitFollowupReq) (*review.Empty, error) {
	// todo: add your logic here and delete this line

	return &review.Empty{}, nil
}
