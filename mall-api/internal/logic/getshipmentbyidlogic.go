// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShipmentByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetShipmentByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShipmentByIdLogic {
	return &GetShipmentByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetShipmentByIdLogic) GetShipmentById(req *types.GetShipmentByIdReq) (*types.GetShipmentByIdResp, error) {
	s, err := l.svcCtx.LogisticsRpc.GetShipment(l.ctx, &logisticspb.GetShipmentReq{ShipmentId: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.GetShipmentByIdResp{Shipment: protoShipmentToType(s)}, nil
}
