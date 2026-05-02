package logic

import (
	"context"

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
	return &ListShipmentsByOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShipmentsByOrderLogic) ListShipmentsByOrder(in *logistics.ListShipmentsByOrderReq) (*logistics.ListShipmentsByOrderResp, error) {
	// todo: add your logic here and delete this line

	return &logistics.ListShipmentsByOrderResp{}, nil
}
