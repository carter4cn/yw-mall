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

type ListRuleSetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRuleSetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRuleSetsLogic {
	return &ListRuleSetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRuleSetsLogic) ListRuleSets(in *rule.ListRuleSetsReq) (*rule.ListRuleSetsResp, error) {
	page, size := in.Page, in.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	rows := make([]model.RuleSet, 0)
	q := fmt.Sprintf(
		"SELECT `id`, `code`, `op`, `member_rule_ids`, `description`, `create_time`, `update_time` FROM `rule_set` ORDER BY id DESC LIMIT %d OFFSET %d",
		size, offset)
	if err := rawQuery(l.ctx, l.svcCtx, q, &rows); err != nil {
		return nil, err
	}
	var total int64
	if err := rawCount(l.ctx, l.svcCtx, "SELECT COUNT(*) FROM `rule_set`", &total); err != nil {
		return nil, err
	}
	out := make([]*rule.RuleSet, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		var ids []int64
		if r.MemberRuleIds != "" {
			_ = json.Unmarshal([]byte(r.MemberRuleIds), &ids)
		}
		out = append(out, &rule.RuleSet{
			Id:            int64(r.Id),
			Code:          r.Code,
			Op:            r.Op,
			MemberRuleIds: ids,
			Description:   r.Description,
		})
	}
	return &rule.ListRuleSetsResp{RuleSets: out, Total: total}, nil
}
