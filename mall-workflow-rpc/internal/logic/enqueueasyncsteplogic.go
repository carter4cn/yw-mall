package logic

import (
	"context"
	"time"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

type EnqueueAsyncStepLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnqueueAsyncStepLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnqueueAsyncStepLogic {
	return &EnqueueAsyncStepLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnqueueAsyncStepLogic) EnqueueAsyncStep(in *workflow.EnqueueAsyncStepReq) (*workflow.Empty, error) {
	opts := []asynq.Option{}
	if in.MaxRetry > 0 {
		opts = append(opts, asynq.MaxRetry(int(in.MaxRetry)))
	}
	if in.TimeoutSeconds > 0 {
		opts = append(opts, asynq.Timeout(time.Duration(in.TimeoutSeconds)*time.Second))
	}
	if in.Queue != "" {
		opts = append(opts, asynq.Queue(in.Queue))
	}
	task := asynq.NewTask(in.TaskType, []byte(in.PayloadJson))
	if _, err := l.svcCtx.AsynqClient.EnqueueContext(l.ctx, task, opts...); err != nil {
		return nil, err
	}
	return &workflow.Empty{}, nil
}
