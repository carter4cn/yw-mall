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

type ListOrderShipmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrderShipmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderShipmentsLogic {
	return &ListOrderShipmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrderShipmentsLogic) ListOrderShipments(req *types.ListOrderShipmentsReq) (*types.ListOrderShipmentsResp, error) {
	r, err := l.svcCtx.LogisticsRpc.ListShipmentsByOrder(l.ctx, &logisticspb.ListShipmentsByOrderReq{OrderId: req.Id})
	if err != nil {
		return nil, err
	}
	out := make([]types.ShipmentDTO, 0, len(r.Shipments))
	for _, s := range r.Shipments {
		out = append(out, protoShipmentToType(s))
	}
	return &types.ListOrderShipmentsResp{Shipments: out}, nil
}
