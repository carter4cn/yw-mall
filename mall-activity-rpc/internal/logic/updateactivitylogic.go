package logic

import (
	"context"
	"database/sql"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityLogic {
	return &UpdateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateActivityLogic) UpdateActivity(in *activity.UpdateActivityReq) (*activity.Empty, error) {
	a, err := l.svcCtx.ActivityModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	if in.Title != "" {
		a.Title = in.Title
	}
	if in.Description != "" {
		a.Description = sql.NullString{String: in.Description, Valid: true}
	}
	if in.StartTime > 0 {
		a.StartTime = time.Unix(in.StartTime, 0)
	}
	if in.EndTime > 0 {
		a.EndTime = time.Unix(in.EndTime, 0)
	}
	if in.ConfigJson != "" {
		a.ConfigJson = sql.NullString{String: in.ConfigJson, Valid: true}
	}
	if err := l.svcCtx.ActivityModel.Update(l.ctx, a); err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
