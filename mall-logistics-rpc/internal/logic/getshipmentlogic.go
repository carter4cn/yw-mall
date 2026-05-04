package logic

import (
	"context"
	"errors"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShipmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShipmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShipmentLogic {
	return &GetShipmentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetShipmentLogic) GetShipment(in *logistics.GetShipmentReq) (*logistics.Shipment, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errors.New("logistics: shipment not found")
	}
	items, _ := fetchItems(l.ctx, l.svcCtx, s.Id)
	tracks, _ := fetchTracks(l.ctx, l.svcCtx, s.Id)
	return toShipmentProto(s, items, tracks), nil
}
