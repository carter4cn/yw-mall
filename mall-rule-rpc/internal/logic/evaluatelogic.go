package logic

import (
	"context"
	"time"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type EvaluateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEvaluateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EvaluateLogic {
	return &EvaluateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EvaluateLogic) Evaluate(in *rule.EvaluateReq) (*rule.EvaluateResp, error) {
	start := time.Now()
	prog, _, err := l.svcCtx.Loader.LoadById(l.ctx, in.RuleId)
	if err != nil {
		return nil, err
	}
	res, err := engine.Run(prog, in.Context)
	latency := time.Since(start).Microseconds()
	if err != nil {
		return &rule.EvaluateResp{Result: false, Detail: err.Error(), LatencyUs: latency}, nil
	}
	return &rule.EvaluateResp{Result: res, LatencyUs: latency}, nil
}
