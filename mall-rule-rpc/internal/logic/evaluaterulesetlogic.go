package logic

import (
	"context"
	"encoding/json"
	"time"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type EvaluateRuleSetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEvaluateRuleSetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EvaluateRuleSetLogic {
	return &EvaluateRuleSetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EvaluateRuleSetLogic) EvaluateRuleSet(in *rule.EvaluateRuleSetReq) (*rule.EvaluateRuleSetResp, error) {
	start := time.Now()
	rs, err := l.svcCtx.RuleSetModel.FindOne(l.ctx, uint64(in.RuleSetId))
	if err != nil {
		return nil, err
	}
	var memberIds []int64
	if err := json.Unmarshal([]byte(rs.MemberRuleIds), &memberIds); err != nil {
		return nil, err
	}

	var firstFailed int64
	switch rs.Op {
	case "AND":
		for _, id := range memberIds {
			ok, err := l.evalOne(id, in.Context)
			if err != nil || !ok {
				return &rule.EvaluateRuleSetResp{
					Result:            false,
					FirstFailedRuleId: id,
					LatencyUs:         time.Since(start).Microseconds(),
				}, nil
			}
		}
		return &rule.EvaluateRuleSetResp{Result: true, LatencyUs: time.Since(start).Microseconds()}, nil

	case "OR":
		anyPass := false
		for _, id := range memberIds {
			ok, err := l.evalOne(id, in.Context)
			if err == nil && ok {
				anyPass = true
				break
			}
			if firstFailed == 0 {
				firstFailed = id
			}
		}
		if !anyPass {
			return &rule.EvaluateRuleSetResp{Result: false, FirstFailedRuleId: firstFailed, LatencyUs: time.Since(start).Microseconds()}, nil
		}
		return &rule.EvaluateRuleSetResp{Result: true, LatencyUs: time.Since(start).Microseconds()}, nil

	case "NOT":
		if len(memberIds) != 1 {
			return nil, errInvalidNotMembers
		}
		ok, err := l.evalOne(memberIds[0], in.Context)
		if err != nil {
			return nil, err
		}
		return &rule.EvaluateRuleSetResp{Result: !ok, LatencyUs: time.Since(start).Microseconds()}, nil
	}
	return nil, errUnknownOp
}

func (l *EvaluateRuleSetLogic) evalOne(ruleId int64, ctx *rule.RuleContext) (bool, error) {
	prog, _, err := l.svcCtx.Loader.LoadById(l.ctx, ruleId)
	if err != nil {
		return false, err
	}
	return engine.Run(prog, ctx)
}

var (
	errInvalidNotMembers = &ruleSetError{msg: "NOT op requires exactly one member rule"}
	errUnknownOp         = &ruleSetError{msg: "unknown op"}
)

type ruleSetError struct{ msg string }

func (e *ruleSetError) Error() string { return e.msg }
