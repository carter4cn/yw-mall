package logic

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DispatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DispatchLogic {
	return &DispatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// envelope is the cross-service contract carried on the Kafka outbox topics.
// Consumers (per-type dispatchers in P5) re-derive idempotency from RecordId+Key.
type envelope struct {
	RecordId           int64  `json:"record_id"`
	UserId             int64  `json:"user_id"`
	ActivityId         int64  `json:"activity_id"`
	WorkflowInstanceId int64  `json:"workflow_instance_id"`
	TemplateId         int64  `json:"template_id"`
	Type               string `json:"type"`
	Payload            string `json:"payload_json"`
	IdempotencyKey     string `json:"idempotency_key"`
}

// Dispatch writes reward_record + outbox in one local tx. The unique index on
// reward_record.idempotency_key collapses concurrent retries to a single row;
// the matching outbox row is the durable signal for the relay → Kafka path.
//
// Topic naming follows the plan: reward.dispatch.<type>.
func (l *DispatchLogic) Dispatch(in *reward.DispatchReq) (*reward.DispatchResp, error) {
	idem := in.IdempotencyKey
	if idem == "" {
		// Caller didn't supply one — derive a deterministic key so retries collapse.
		// Workflow-instance-scoped: the same workflow can't double-dispatch the same template.
		h := sha1.Sum([]byte(fmt.Sprintf("%d|%d|%d|%d", in.UserId, in.ActivityId, in.WorkflowInstanceId, in.TemplateId)))
		idem = hex.EncodeToString(h[:])
	}

	// Fast path: unique-index lookup (cached). Skips the tx for the steady-state retry case.
	if existing, err := l.svcCtx.RewardRecordModel.FindOneByIdempotencyKey(l.ctx, idem); err == nil && existing != nil {
		return &reward.DispatchResp{RewardRecordId: int64(existing.Id), Status: existing.Status}, nil
	}

	tmpl, err := l.svcCtx.RewardTemplateModel.FindOne(l.ctx, uint64(in.TemplateId))
	if err != nil {
		return nil, fmt.Errorf("template %d: %w", in.TemplateId, err)
	}
	if tmpl.Status != "ACTIVE" {
		return nil, fmt.Errorf("template %d not active (status=%s)", in.TemplateId, tmpl.Status)
	}

	var newRecordId int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// Re-check inside the tx with FOR UPDATE — closes the race where two callers
		// arrive concurrently after the cached fast-path miss.
		var existingId int64
		if err := session.QueryRowCtx(ctx, &existingId,
			"SELECT id FROM `reward_record` WHERE idempotency_key=? LIMIT 1 FOR UPDATE", idem); err != nil && err != sqlx.ErrNotFound {
			return err
		}
		if existingId > 0 {
			newRecordId = existingId
			return nil
		}

		res, err := session.ExecCtx(ctx,
			"INSERT INTO `reward_record`(user_id, activity_id, workflow_instance_id, template_id, `type`, payload_json, status, idempotency_key, version) VALUES (?,?,?,?,?,?,?,?,0)",
			in.UserId, in.ActivityId, in.WorkflowInstanceId, in.TemplateId, tmpl.Type,
			sql.NullString{String: in.PayloadJson, Valid: in.PayloadJson != ""},
			"PENDING", idem,
		)
		if err != nil {
			return err
		}
		newRecordId, _ = res.LastInsertId()

		envBytes, _ := json.Marshal(envelope{
			RecordId:           newRecordId,
			UserId:             in.UserId,
			ActivityId:         in.ActivityId,
			WorkflowInstanceId: in.WorkflowInstanceId,
			TemplateId:         in.TemplateId,
			Type:               tmpl.Type,
			Payload:            in.PayloadJson,
			IdempotencyKey:     idem,
		})

		topic := "reward.dispatch." + tmpl.Type
		_, err = session.ExecCtx(ctx,
			"INSERT INTO `outbox`(topic, `key`, payload, status) VALUES (?,?,?,?)",
			topic, fmt.Sprintf("%d", newRecordId), string(envBytes), "PENDING",
		)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &reward.DispatchResp{RewardRecordId: newRecordId, Status: "PENDING"}, nil
}
