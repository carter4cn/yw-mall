package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePreOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePreOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePreOrderLogic {
	return &CreatePreOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePreOrderLogic) CreatePreOrder(in *order.CreatePreOrderReq) (*order.CreatePreOrderResp, error) {
	// todo: add your logic here and delete this line

	return &order.CreatePreOrderResp{}, nil
}
