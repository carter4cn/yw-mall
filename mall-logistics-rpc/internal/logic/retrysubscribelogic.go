package logic

import (
	"context"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrySubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySubscribeLogic {
	return &RetrySubscribeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetrySubscribeLogic) RetrySubscribe(in *logistics.RetrySubscribeReq) (*logistics.Empty, error) {
	// todo: add your logic here and delete this line

	return &logistics.Empty{}, nil
}
