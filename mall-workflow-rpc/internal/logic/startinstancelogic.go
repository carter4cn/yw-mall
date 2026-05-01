package logic

import (
	"context"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartInstanceLogic {
	return &StartInstanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartInstanceLogic) StartInstance(in *workflow.StartInstanceReq) (*workflow.StartInstanceResp, error) {
	def, err := l.svcCtx.Registry.Get(l.ctx, uint64(in.DefinitionId))
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.Persister.CreateInstance(l.ctx, in.DefinitionId, in.ActivityId, in.UserId, def.InitialState, in.PayloadJson)
	if err != nil {
		return nil, err
	}
	return &workflow.StartInstanceResp{InstanceId: id, State: def.InitialState}, nil
}
