package logic

import (
	"context"
	"fmt"
	"time"

	"mall-workflow-rpc/internal/fsm"
	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type FireLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFireLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FireLogic {
	return &FireLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FireLogic) Fire(in *workflow.FireReq) (*workflow.FireResp, error) {
	inst, err := l.svcCtx.Persister.LoadInstance(l.ctx, uint64(in.InstanceId))
	if err != nil {
		return nil, err
	}
	def, err := l.svcCtx.Registry.Get(l.ctx, uint64(inst.DefinitionId))
	if err != nil {
		return nil, err
	}

	sm := fsm.Build(def, inst.State)
	start := time.Now()
	fireErr := sm.Fire(in.Trigger)
	latency := time.Since(start)

	state, _ := sm.State(l.ctx)
	to, _ := state.(string)

	if fireErr != nil {
		_ = l.svcCtx.Persister.Apply(l.ctx, uint64(inst.Id), inst.State, inst.State, in.Trigger, latency, in.PayloadJson, fireErr)
		return nil, fmt.Errorf("fire trigger=%s on state=%s: %w", in.Trigger, inst.State, fireErr)
	}

	advanced := to != inst.State
	if advanced {
		if err := l.svcCtx.Persister.Apply(l.ctx, uint64(inst.Id), inst.State, to, in.Trigger, latency, in.PayloadJson, nil); err != nil {
			return nil, err
		}
	}
	return &workflow.FireResp{State: to, Advanced: advanced}, nil
}
