package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type IngestWebhookEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIngestWebhookEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IngestWebhookEventsLogic {
	return &IngestWebhookEventsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *IngestWebhookEventsLogic) IngestWebhookEvents(in *logistics.IngestWebhookEventsReq) (*logistics.Empty, error) {
	if in.TrackingNo == "" || in.Carrier == "" {
		return nil, errors.New("logistics: tracking_no and carrier are required")
	}
	var ship struct {
		Id     int64 `db:"id"`
		Status int64 `db:"status"`
	}
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &ship,
		"SELECT id, status FROM shipment WHERE carrier=? AND tracking_no=? LIMIT 1",
		in.Carrier, in.TrackingNo); err != nil {
		if err == sql.ErrNoRows || err == sqlx.ErrNotFound {
			return nil, errors.New("logistics: shipment not found")
		}
		return nil, err
	}

	var maxInternal int32
	var lastTime time.Time
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, e := range in.Events {
			t := time.Unix(e.TrackTime, 0)
			_, err := session.ExecCtx(ctx,
				"INSERT IGNORE INTO `shipment_track`(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?,?,?,?,?,?)",
				ship.Id, t, e.Location, e.Description, e.StateKuaidi100, e.StateInternal)
			if err != nil {
				return err
			}
			if e.StateInternal > maxInternal {
				maxInternal = e.StateInternal
			}
			if t.After(lastTime) {
				lastTime = t
			}
		}
		_, err := session.ExecCtx(ctx,
			"UPDATE `shipment` SET status=GREATEST(status,?), last_track_time=GREATEST(IFNULL(last_track_time,'1970-01-01'),?), subscribe_status=1 WHERE id=?",
			maxInternal, lastTime, ship.Id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
