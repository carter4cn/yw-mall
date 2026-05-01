package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"mall-rule-rpc/internal/model"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRuleSetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRuleSetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRuleSetLogic {
	return &CreateRuleSetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRuleSetLogic) CreateRuleSet(in *rule.CreateRuleSetReq) (*rule.CreateRuleSetResp, error) {
	op := in.Op
	switch op {
	case "AND", "OR", "NOT":
	case "":
		op = "AND"
	default:
		return nil, fmt.Errorf("invalid op %q (allowed: AND/OR/NOT)", op)
	}
	if op == "NOT" && len(in.MemberRuleIds) != 1 {
		return nil, fmt.Errorf("NOT requires exactly one member rule")
	}
	idsJson, err := json.Marshal(in.MemberRuleIds)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.RuleSetModel.Insert(l.ctx, &model.RuleSet{
		Code:          in.Code,
		Op:            op,
		MemberRuleIds: string(idsJson),
		Description:   in.Description,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &rule.CreateRuleSetResp{Id: id}, nil
}
