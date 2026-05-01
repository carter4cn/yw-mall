package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordEventLogic {
	return &RecordEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecordEventLogic) RecordEvent(in *risk.RecordEventReq) (*risk.Empty, error) {
	// todo: add your logic here and delete this line

	return &risk.Empty{}, nil
}
