package logic

import (
	"context"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchDispatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDispatchLogic {
	return &BatchDispatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchDispatchLogic) BatchDispatch(in *reward.BatchDispatchReq) (*reward.BatchDispatchResp, error) {
	out := &reward.BatchDispatchResp{Results: make([]*reward.DispatchResp, 0, len(in.Items))}
	for _, item := range in.Items {
		res, err := NewDispatchLogic(l.ctx, l.svcCtx).Dispatch(item)
		if err != nil {
			// per-item failure mode: return what succeeded; caller retries with same idempotency keys
			out.Results = append(out.Results, &reward.DispatchResp{Status: "FAILED"})
			continue
		}
		out.Results = append(out.Results, res)
	}
	return out, nil
}
