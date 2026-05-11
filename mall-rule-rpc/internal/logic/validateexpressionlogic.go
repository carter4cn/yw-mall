package logic

import (
	"context"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateExpressionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateExpressionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateExpressionLogic {
	return &ValidateExpressionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateExpression compiles the expression with the rule engine and reports
// any syntactic or type error without persisting anything.
func (l *ValidateExpressionLogic) ValidateExpression(in *rule.ValidateExpressionReq) (*rule.ValidateExpressionResp, error) {
	if _, err := engine.Compile(in.Expression); err != nil {
		return &rule.ValidateExpressionResp{Valid: false, Error: err.Error()}, nil
	}
	return &rule.ValidateExpressionResp{Valid: true}, nil
}
