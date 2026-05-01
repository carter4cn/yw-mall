package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PauseActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPauseActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseActivityLogic {
	return &PauseActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PauseActivityLogic) PauseActivity(in *activity.IdReq) (*activity.Empty, error) {
	a, err := l.svcCtx.ActivityModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	a.Status = "PAUSED"
	if err := l.svcCtx.ActivityModel.Update(l.ctx, a); err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
