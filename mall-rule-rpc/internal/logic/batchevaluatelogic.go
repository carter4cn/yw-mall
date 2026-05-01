package logic

import (
	"context"

	"mall-rule-rpc/internal/engine"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchEvaluateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchEvaluateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchEvaluateLogic {
	return &BatchEvaluateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchEvaluateLogic) BatchEvaluate(in *rule.BatchEvaluateReq) (*rule.BatchEvaluateResp, error) {
	results := make(map[int64]bool, len(in.RuleIds))
	var firstFailed int64
	for _, id := range in.RuleIds {
		prog, _, err := l.svcCtx.Loader.LoadById(l.ctx, id)
		if err != nil {
			results[id] = false
			if firstFailed == 0 {
				firstFailed = id
			}
			continue
		}
		ok, err := engine.Run(prog, in.Context)
		if err != nil || !ok {
			results[id] = false
			if firstFailed == 0 {
				firstFailed = id
			}
			continue
		}
		results[id] = true
	}
	return &rule.BatchEvaluateResp{Results: results, FirstFailed: firstFailed}, nil
}
