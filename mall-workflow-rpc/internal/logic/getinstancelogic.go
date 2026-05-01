package logic

import (
	"context"
	"time"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInstanceLogic {
	return &GetInstanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type stepRow struct {
	Id         int64     `db:"id"`
	InstanceId int64     `db:"instance_id"`
	FromState  string    `db:"from_state"`
	ToState    string    `db:"to_state"`
	Trigger    string    `db:"trigger"`
	LatencyMs  int64     `db:"latency_ms"`
	Error      string    `db:"error"`
	CreateTime time.Time `db:"create_time"`
}

func (l *GetInstanceLogic) GetInstance(in *workflow.IdReq) (*workflow.GetInstanceResp, error) {
	inst, err := l.svcCtx.Persister.LoadInstance(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	resp := &workflow.GetInstanceResp{
		Instance: &workflow.WorkflowInstance{
			Id:           int64(inst.Id),
			DefinitionId: inst.DefinitionId,
			ActivityId:   inst.ActivityId,
			UserId:       inst.UserId,
			State:        inst.State,
			PayloadJson:  inst.PayloadJson.String,
			Version:      int32(inst.Version),
			LastEventAt:  inst.LastEventAt.Unix(),
			CreateTime:   inst.CreateTime.Unix(),
		},
	}
	rows := []stepRow{}
	q := "SELECT id, instance_id, from_state, to_state, `trigger`, latency_ms, error, create_time FROM `workflow_step_log` WHERE instance_id=? ORDER BY id ASC LIMIT 200"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, inst.Id); err == nil {
		for i := range rows {
			r := &rows[i]
			resp.Steps = append(resp.Steps, &workflow.StepLog{
				Id:         r.Id,
				InstanceId: r.InstanceId,
				FromState:  r.FromState,
				ToState:    r.ToState,
				Trigger:    r.Trigger,
				LatencyMs:  r.LatencyMs,
				Error:      r.Error,
				CreateTime: r.CreateTime.Unix(),
			})
		}
	}
	return resp, nil
}
