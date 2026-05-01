package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EndActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEndActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EndActivityLogic {
	return &EndActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EndActivityLogic) EndActivity(in *activity.IdReq) (*activity.Empty, error) {
	a, err := l.svcCtx.ActivityModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	a.Status = "ENDED"
	if err := l.svcCtx.ActivityModel.Update(l.ctx, a); err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
