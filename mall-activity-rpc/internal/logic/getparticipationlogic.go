package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetParticipationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetParticipationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetParticipationLogic {
	return &GetParticipationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetParticipationLogic) GetParticipation(in *activity.GetParticipationReq) (*activity.ParticipationRecord, error) {
	r, err := l.svcCtx.ParticipationRecordModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &activity.ParticipationRecord{
		Id:                 int64(r.Id),
		ActivityId:         r.ActivityId,
		UserId:             r.UserId,
		Sequence:           r.Sequence,
		WorkflowInstanceId: r.WorkflowInstanceId,
		Status:             r.Status,
		PayloadJson:        r.PayloadJson.String,
		CreateTime:         r.CreateTime.Unix(),
	}, nil
}
