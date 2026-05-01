package logic

import (
	"context"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTemplateLogic {
	return &GetTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTemplateLogic) GetTemplate(in *reward.IdReq) (*reward.RewardTemplate, error) {
	t, err := l.svcCtx.RewardTemplateModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &reward.RewardTemplate{
		Id:                int64(t.Id),
		Code:              t.Code,
		Type:              t.Type,
		PayloadSchemaJson: t.PayloadSchemaJson.String,
		MaxValue:          t.MaxValue,
		Status:            t.Status,
		Description:       t.Description,
	}, nil
}
