package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/model"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateActivityRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateActivityRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityRuleLogic {
	return &CreateActivityRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateActivityRule turns the high-level low-code definition (conditions /
// exclusions / rewards) into a single expr-lang rule plus a wrapper rule_set
// pointing at it. The generated expression is also returned for transparency.
func (l *CreateActivityRuleLogic) CreateActivityRule(in *rule.CreateActivityRuleReq) (*rule.CreateActivityRuleResp, error) {
	expression := buildExpression(in.Conditions, in.Exclusions)
	if expression == "" {
		expression = "true"
	}
	if _, err := engine.Compile(expression); err != nil {
		return nil, fmt.Errorf("generated expression is invalid: %w", err)
	}

	jsonSchema := buildSchema(in)

	rRes, err := l.svcCtx.RuleModel.Insert(l.ctx, &model.Rule{
		Code:        in.Code,
		Description: in.Description,
		Expression:  expression,
		Lang:        "expr",
		Version:     1,
		Status:      "ACTIVE",
		JsonSchema:  sql.NullString{String: jsonSchema, Valid: jsonSchema != ""},
	})
	if err != nil {
		return nil, err
	}
	ruleId, _ := rRes.LastInsertId()

	idsJson, _ := json.Marshal([]int64{ruleId})
	sRes, err := l.svcCtx.RuleSetModel.Insert(l.ctx, &model.RuleSet{
		Code:          in.Code + "_set",
		Op:            "AND",
		MemberRuleIds: string(idsJson),
		Description:   in.Description,
	})
	if err != nil {
		return nil, err
	}
	setId, _ := sRes.LastInsertId()

	return &rule.CreateActivityRuleResp{
		RuleId:              ruleId,
		RuleSetId:           setId,
		GeneratedExpression: expression,
	}, nil
}

// buildExpression converts conditions and exclusions into an expr-lang clause:
//
//	(cond1 && cond2 && !excl1 && !excl2)
//
// Unknown types degrade to `true` so they do not block evaluation.
func buildExpression(conds []*rule.ActivityRuleCondition, excls []*rule.ActivityRuleExclusion) string {
	parts := []string{}
	for _, c := range conds {
		if s := condToExpr(c); s != "" {
			parts = append(parts, s)
		}
	}
	for _, e := range excls {
		if s := exclToExpr(e); s != "" {
			parts = append(parts, "!("+s+")")
		}
	}
	return strings.Join(parts, " && ")
}

func condToExpr(c *rule.ActivityRuleCondition) string {
	if c == nil {
		return ""
	}
	op := opSym(c.Operator)
	if op == "" {
		return ""
	}
	switch c.Type {
	case "order_amount", "signin_count", "activity_count", "participation_count_today", "participation_count_total":
		return fmt.Sprintf("%s %s %d", c.Type, op, c.Value)
	}
	return ""
}

func exclToExpr(e *rule.ActivityRuleExclusion) string {
	if e == nil {
		return ""
	}
	switch e.Type {
	case "blacklist":
		return "is_blacklisted"
	case "risk_label":
		return "risk_score >= 60"
	case "holiday":
		return fmt.Sprintf("holiday_type == %q", e.Value)
	}
	return ""
}

func opSym(op string) string {
	switch op {
	case "gte", "":
		return ">="
	case "lte":
		return "<="
	case "eq":
		return "=="
	case "gt":
		return ">"
	case "lt":
		return "<"
	}
	return ""
}

// buildSchema embeds the original low-code spec so it can round-trip through
// the admin UI. We don't try to make it a real JSON schema.
func buildSchema(in *rule.CreateActivityRuleReq) string {
	payload := map[string]any{
		"budget":     in.Budget,
		"conditions": in.Conditions,
		"exclusions": in.Exclusions,
		"rewards":    in.Rewards,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(b)
}
