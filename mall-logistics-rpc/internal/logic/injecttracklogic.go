package logic

import (
	"context"
	"errors"
	"time"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type InjectTrackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInjectTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InjectTrackLogic {
	return &InjectTrackLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *InjectTrackLogic) InjectTrack(in *logistics.InjectTrackReq) (*logistics.Empty, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errors.New("logistics: shipment not found")
	}
	now := time.Now()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx,
			"INSERT IGNORE INTO `shipment_track`(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?,?,?,?,NULL,?)",
			s.Id, now, in.Location, in.Description, in.StateInternal); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx,
			"UPDATE `shipment` SET status=GREATEST(status,?), last_track_time=? WHERE id=?",
			in.StateInternal, now, s.Id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
