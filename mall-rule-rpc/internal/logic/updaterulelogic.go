package logic

import (
	"context"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRuleLogic {
	return &UpdateRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRuleLogic) UpdateRule(in *rule.UpdateRuleReq) (*rule.Empty, error) {
	r, err := l.svcCtx.RuleModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	if in.Expression != "" && in.Expression != r.Expression {
		if _, err := engine.Compile(in.Expression); err != nil {
			return nil, err
		}
		r.Expression = in.Expression
		r.Version++
	}
	if in.Description != "" {
		r.Description = in.Description
	}
	if in.Status != "" {
		r.Status = in.Status
	}
	if err := l.svcCtx.RuleModel.Update(l.ctx, r); err != nil {
		return nil, err
	}
	return &rule.Empty{}, nil
}
