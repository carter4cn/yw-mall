package logic

import (
	"context"
	"database/sql"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/model"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRuleLogic {
	return &CreateRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRuleLogic) CreateRule(in *rule.CreateRuleReq) (*rule.CreateRuleResp, error) {
	if _, err := engine.Compile(in.Expression); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.RuleModel.Insert(l.ctx, &model.Rule{
		Code:        in.Code,
		Description: in.Description,
		Expression:  in.Expression,
		Lang:        "expr",
		Version:     1,
		Status:      "ACTIVE",
		JsonSchema:  sql.NullString{String: in.JsonSchema, Valid: in.JsonSchema != ""},
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &rule.CreateRuleResp{Id: id}, nil
}
