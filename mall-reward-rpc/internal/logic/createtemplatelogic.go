package logic

import (
	"context"
	"database/sql"

	"mall-reward-rpc/internal/model"
	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTemplateLogic {
	return &CreateTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateTemplate is idempotent on `code` so the seed binary can be re-run safely.
func (l *CreateTemplateLogic) CreateTemplate(in *reward.CreateTemplateReq) (*reward.CreateTemplateResp, error) {
	if existing, err := l.svcCtx.RewardTemplateModel.FindOneByCode(l.ctx, in.Code); err == nil && existing != nil {
		return &reward.CreateTemplateResp{Id: int64(existing.Id)}, nil
	}
	res, err := l.svcCtx.RewardTemplateModel.Insert(l.ctx, &model.RewardTemplate{
		Code:              in.Code,
		Type:              in.Type,
		PayloadSchemaJson: sql.NullString{String: in.PayloadSchemaJson, Valid: in.PayloadSchemaJson != ""},
		MaxValue:          in.MaxValue,
		Status:            "ACTIVE",
		Description:       in.Description,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &reward.CreateTemplateResp{Id: id}, nil
}
