package logic

import (
	"context"

	"mall-logistics-rpc/internal/model"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShipmentsByOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShipmentsByOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShipmentsByOrderLogic {
	return &ListShipmentsByOrderLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListShipmentsByOrderLogic) ListShipmentsByOrder(in *logistics.ListShipmentsByOrderReq) (*logistics.ListShipmentsByOrderResp, error) {
	rows := []*model.Shipment{}
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, order_id, user_id, tracking_no, carrier, status, subscribe_status, last_track_time, create_time, update_time FROM shipment WHERE order_id=? ORDER BY create_time",
		in.OrderId); err != nil {
		return nil, err
	}
	out := make([]*logistics.Shipment, 0, len(rows))
	for _, s := range rows {
		items, _ := fetchItems(l.ctx, l.svcCtx, s.Id)
		tracks, _ := fetchTracks(l.ctx, l.svcCtx, s.Id)
		out = append(out, toShipmentProto(s, items, tracks))
	}
	return &logistics.ListShipmentsByOrderResp{Shipments: out}, nil
}
