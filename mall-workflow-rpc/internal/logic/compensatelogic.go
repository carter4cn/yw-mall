package logic

import (
	"context"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompensateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompensateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompensateLogic {
	return &CompensateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CompensateLogic) Compensate(in *workflow.CompensateReq) (*workflow.Empty, error) {
	if err := l.svcCtx.Persister.ForceState(l.ctx, uint64(in.InstanceId), "COMPENSATED", in.Reason); err != nil {
		return nil, err
	}
	return &workflow.Empty{}, nil
}
