package logic

import (
	"context"

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
	return &GetShipmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetShipmentLogic) GetShipment(in *logistics.GetShipmentReq) (*logistics.Shipment, error) {
	// todo: add your logic here and delete this line

	return &logistics.Shipment{}, nil
}
