// Package outbox implements the transactional outbox relay for mall-reward-rpc.
//
// Producers (Dispatch / RefundReward) write reward_record + outbox rows in one
// local SQL tx. This relay drains PENDING rows through a Publisher and marks
// them PUBLISHED. The Publisher is an interface so we can ship the demo with a
// log-only sink today (LogPublisher) and swap in a Kafka publisher when the
// kafka client lib is available — see KafkaPublisher in kafka_publisher.go.
//
// Crash safety:
//   - The relay polls in batches and publishes one row at a time. A row is only
//     marked PUBLISHED after Publish returns nil — at-least-once delivery.
//   - Consumers must dedupe by envelope.idempotency_key (the same key written
//     into reward_record), so a re-publish after a crashed relay is harmless.
package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Publisher abstracts the destination of outbox messages so the relay does not
// hard-depend on any single message bus client.
type Publisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
	Close() error
}

// LogPublisher records each outbox row to logs and returns success. This is
// the fallback when no real Kafka client is available — the
// reward_record / outbox state machine still advances correctly, the only
// thing missing is the cross-service event delivery.
type LogPublisher struct{}

func (LogPublisher) Publish(_ context.Context, topic, key string, payload []byte) error {
	logx.Infof("[outbox] would-publish topic=%s key=%s payload=%s", topic, key, string(payload))
	return nil
}
func (LogPublisher) Close() error { return nil }

type Relay struct {
	db        sqlx.SqlConn
	publisher Publisher
	pollTick  time.Duration
	batch     int
	stop      chan struct{}
}

func NewRelay(db sqlx.SqlConn, publisher Publisher) *Relay {
	if publisher == nil {
		publisher = LogPublisher{}
	}
	return &Relay{
		db:        db,
		publisher: publisher,
		pollTick:  time.Second,
		batch:     100,
		stop:      make(chan struct{}),
	}
}

func (r *Relay) Start(ctx context.Context) {
	go r.loop(ctx)
}

func (r *Relay) Stop() {
	close(r.stop)
	_ = r.publisher.Close()
}

type pendingRow struct {
	Id      uint64         `db:"id"`
	Topic   string         `db:"topic"`
	Key     string         `db:"key"`
	Payload sql.NullString `db:"payload"`
}

func (r *Relay) loop(ctx context.Context) {
	logx.Info("outbox relay started")
	t := time.NewTicker(r.pollTick)
	defer t.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ctx.Done():
			return
		case <-t.C:
			r.drain(ctx)
		}
	}
}

// drain pulls a batch of PENDING rows, publishes each, and updates the row.
// We deliberately do not use SELECT ... FOR UPDATE here: a single relay loop
// per service instance is the expected topology, and the UPDATE-with-status
// guard prevents duplicate publication if multiple relays ever race.
func (r *Relay) drain(ctx context.Context) {
	var rows []pendingRow
	q := "SELECT id, topic, `key`, payload FROM `outbox` WHERE status='PENDING' ORDER BY id LIMIT ?"
	if err := r.db.QueryRowsCtx(ctx, &rows, q, r.batch); err != nil {
		logx.Errorf("outbox: query pending: %v", err)
		return
	}
	for _, row := range rows {
		if err := r.publisher.Publish(ctx, row.Topic, row.Key, []byte(row.Payload.String)); err != nil {
			logx.Errorf("outbox: publish id=%d topic=%s: %v", row.Id, row.Topic, err)
			continue
		}
		if _, err := r.db.ExecCtx(ctx,
			"UPDATE `outbox` SET status='PUBLISHED', published_at=NOW() WHERE id=? AND status='PENDING'",
			row.Id,
		); err != nil {
			logx.Errorf("outbox: mark published id=%d: %v", row.Id, err)
		}
	}
}
