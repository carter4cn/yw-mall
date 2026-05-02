package logic

import (
	"context"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShipmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShipmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShipmentLogic {
	return &CreateShipmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateShipmentLogic) CreateShipment(in *logistics.CreateShipmentReq) (*logistics.CreateShipmentResp, error) {
	// todo: add your logic here and delete this line

	return &logistics.CreateShipmentResp{}, nil
}
