package logic

import (
	"context"
	"fmt"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishActivityLogic {
	return &PublishActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PublishActivity flips status DRAFT→PUBLISHED and, for seckill activities,
// preheats the inventory snapshots into Redis so the hot path can decrement
// atomically without touching MySQL.
//
// Reads use the write conn + FOR UPDATE so ProxySQL routes to master,
// avoiding stale-replica reads when the row was just inserted.
func (l *PublishActivityLogic) PublishActivity(in *activity.IdReq) (*activity.Empty, error) {
	a, err := l.loadActivityForUpdate(in.Id)
	if err != nil {
		return nil, fmt.Errorf("load activity %d: %w", in.Id, err)
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `activity` SET status='PUBLISHED', version=version+1 WHERE id=?",
		in.Id,
	); err != nil {
		return nil, err
	}
	// bust go-zero's cached-model entries so future FindOne sees the new status
	_, _ = l.svcCtx.Redis.DelCtx(l.ctx,
		fmt.Sprintf("cache:activity:id:%d", in.Id),
		fmt.Sprintf("cache:activity:code:%s", a.Code),
	)
	if a.Type == "seckill" {
		if err := l.preheatSeckill(in.Id); err != nil {
			l.Logger.Errorf("preheat seckill activity_id=%d failed: %v", in.Id, err)
		}
	}
	return &activity.Empty{}, nil
}

func (l *PublishActivityLogic) loadActivityForUpdate(id int64) (*activityRow, error) {
	var row activityRow
	q := "SELECT id, code, type, status FROM `activity` WHERE id=? LIMIT 1 FOR UPDATE"
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row, q, id); err != nil {
		return nil, err
	}
	return &row, nil
}

type activityRow struct {
	Id     int64  `db:"id"`
	Code   string `db:"code"`
	Type   string `db:"type"`
	Status string `db:"status"`
}

// preheatSeckill copies activity_inventory_snapshot rows into a Redis hash
// keyed by sku id. Lua scripts decrement these counters atomically.
func (l *PublishActivityLogic) preheatSeckill(activityId int64) error {
	rows := []struct {
		SkuId        int64 `db:"sku_id"`
		CurrentStock int64 `db:"current_stock"`
	}{}
	// FOR UPDATE pins to the writer hostgroup in ProxySQL — without it, this SELECT
	// can race a freshly-inserted snapshot row that hasn't replicated to the slaves yet.
	q := "SELECT sku_id, current_stock FROM `activity_inventory_snapshot` WHERE activity_id=? FOR UPDATE"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, activityId); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	hashKey := fmt.Sprintf("activity:%d:stock", activityId)
	pairs := make(map[string]string, len(rows))
	for _, r := range rows {
		pairs[fmt.Sprintf("%d", r.SkuId)] = fmt.Sprintf("%d", r.CurrentStock)
	}
	if err := l.svcCtx.Redis.HmsetCtx(l.ctx, hashKey, pairs); err != nil {
		return err
	}
	return nil
}
