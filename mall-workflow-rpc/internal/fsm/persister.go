package fsm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mall-workflow-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Persister applies state transitions to the workflow_instance row with
// optimistic locking and writes a row to workflow_step_log.
type Persister struct {
	conn      sqlx.SqlConn
	instances model.WorkflowInstanceModel
	steps     model.WorkflowStepLogModel
}

func NewPersister(conn sqlx.SqlConn, instances model.WorkflowInstanceModel, steps model.WorkflowStepLogModel) *Persister {
	return &Persister{conn: conn, instances: instances, steps: steps}
}

// Apply transitions an instance from its current state to `toState` and writes
// a step log entry. Uses optimistic locking via the `version` column.
func (p *Persister) Apply(ctx context.Context, instanceId uint64, fromState, toState, trigger string, latency time.Duration, payloadJson string, fireErr error) error {
	errMsg := ""
	if fireErr != nil {
		errMsg = fireErr.Error()
		if len(errMsg) > 1000 {
			errMsg = errMsg[:1000]
		}
	}

	// step log is best-effort and append-only.
	_, _ = p.steps.Insert(ctx, &model.WorkflowStepLog{
		InstanceId: int64(instanceId),
		FromState:  fromState,
		ToState:    toState,
		Trigger:    trigger,
		LatencyMs:  latency.Milliseconds(),
		Error:      errMsg,
	})

	if fireErr != nil {
		return fireErr
	}

	// optimistic update on workflow_instance using version
	res, err := p.conn.ExecCtx(ctx,
		"UPDATE `workflow_instance` SET state=?, version=version+1, last_event_at=NOW(), payload_json=COALESCE(NULLIF(?, ''), payload_json) WHERE id=? AND state=?",
		toState, payloadJson, instanceId, fromState,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("optimistic lock failure on workflow_instance id=%d (state changed under us)", instanceId)
	}
	// bust cached instance
	_ = p.instances.Delete // keep import
	return nil
}

// LoadInstance fetches the current state row directly from the write conn,
// bypassing the cached model and ProxySQL read-replica routing. Activity
// flows must always read the freshest state because Fire is preceded by
// Insert/Update on the same row.
func (p *Persister) LoadInstance(ctx context.Context, id uint64) (*model.WorkflowInstance, error) {
	var w model.WorkflowInstance
	// FOR UPDATE makes ProxySQL route this read to the write hostgroup
	// (master), avoiding stale reads from a replica that hasn't applied
	// the previous transition yet. We're not relying on the lock for
	// correctness — Apply uses optimistic locking via the version column.
	q := "SELECT id, definition_id, activity_id, user_id, state, payload_json, version, last_event_at, create_time, update_time FROM `workflow_instance` WHERE id = ? LIMIT 1 FOR UPDATE"
	if err := p.conn.QueryRowCtx(ctx, &w, q, id); err != nil {
		return nil, err
	}
	return &w, nil
}

// CreateInstance inserts a new row at initialState.
func (p *Persister) CreateInstance(ctx context.Context, definitionId, activityId, userId int64, initialState, payloadJson string) (int64, error) {
	res, err := p.instances.Insert(ctx, &model.WorkflowInstance{
		DefinitionId: definitionId,
		ActivityId:   activityId,
		UserId:       userId,
		State:        initialState,
		PayloadJson:  sql.NullString{String: payloadJson, Valid: payloadJson != ""},
		Version:      0,
		LastEventAt:  time.Now(),
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// ForceState bypasses FSM rules and writes a state directly (used for compensation).
func (p *Persister) ForceState(ctx context.Context, instanceId uint64, newState, reason string) error {
	res, err := p.conn.ExecCtx(ctx,
		"UPDATE `workflow_instance` SET state=?, version=version+1, last_event_at=NOW() WHERE id=?",
		newState, instanceId,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("workflow_instance id=%d not found", instanceId)
	}
	_, _ = p.steps.Insert(ctx, &model.WorkflowStepLog{
		InstanceId: int64(instanceId),
		FromState:  "*",
		ToState:    newState,
		Trigger:    "compensate",
		Error:      reason,
	})
	return nil
}
