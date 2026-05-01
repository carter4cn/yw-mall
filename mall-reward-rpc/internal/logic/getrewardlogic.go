package logic

import (
	"context"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRewardLogic {
	return &GetRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRewardLogic) GetReward(in *reward.IdReq) (*reward.RewardRecord, error) {
	r, err := l.svcCtx.RewardRecordModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &reward.RewardRecord{
		Id:                 int64(r.Id),
		UserId:             r.UserId,
		ActivityId:         r.ActivityId,
		WorkflowInstanceId: r.WorkflowInstanceId,
		TemplateId:         r.TemplateId,
		Type:               r.Type,
		PayloadJson:        r.PayloadJson.String,
		Status:             r.Status,
		IdempotencyKey:     r.IdempotencyKey,
		Version:            int32(r.Version),
		CreateTime:         r.CreateTime.Unix(),
		UpdateTime:         r.UpdateTime.Unix(),
	}, nil
}
