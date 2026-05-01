package logic

import (
	"context"

	"mall-workflow-rpc/internal/model"
	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordStepLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordStepLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordStepLogic {
	return &RecordStepLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecordStepLogic) RecordStep(in *workflow.RecordStepReq) (*workflow.Empty, error) {
	if _, err := l.svcCtx.WorkflowStepLogModel.Insert(l.ctx, &model.WorkflowStepLog{
		InstanceId: in.InstanceId,
		FromState:  in.FromState,
		ToState:    in.ToState,
		Trigger:    in.Trigger,
		LatencyMs:  in.LatencyMs,
		Error:      in.Error,
	}); err != nil {
		return nil, err
	}
	return &workflow.Empty{}, nil
}
