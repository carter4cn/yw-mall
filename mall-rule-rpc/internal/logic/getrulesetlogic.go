package logic

import (
	"context"
	"encoding/json"

	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRuleSetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRuleSetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRuleSetLogic {
	return &GetRuleSetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRuleSetLogic) GetRuleSet(in *rule.GetRuleSetReq) (*rule.RuleSet, error) {
	rs, err := l.svcCtx.RuleSetModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	var ids []int64
	if rs.MemberRuleIds != "" {
		_ = json.Unmarshal([]byte(rs.MemberRuleIds), &ids)
	}
	return &rule.RuleSet{
		Id:            int64(rs.Id),
		Code:          rs.Code,
		Op:            rs.Op,
		MemberRuleIds: ids,
		Description:   rs.Description,
	}, nil
}
