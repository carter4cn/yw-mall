package logic

import (
	"context"

	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRuleLogic {
	return &GetRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRuleLogic) GetRule(in *rule.IdReq) (*rule.RuleInfo, error) {
	r, err := l.svcCtx.RuleModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &rule.RuleInfo{
		Id:          int64(r.Id),
		Code:        r.Code,
		Description: r.Description,
		Expression:  r.Expression,
		Lang:        r.Lang,
		Version:     int32(r.Version),
		Status:      r.Status,
		JsonSchema:  r.JsonSchema.String,
		CreateTime:  r.CreateTime.Unix(),
		UpdateTime:  r.UpdateTime.Unix(),
	}, nil
}
