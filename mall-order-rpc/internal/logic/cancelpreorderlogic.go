package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelPreOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelPreOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelPreOrderLogic {
	return &CancelPreOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelPreOrderLogic) CancelPreOrder(in *order.CancelPreOrderReq) (*order.CancelPreOrderResp, error) {
	// todo: add your logic here and delete this line

	return &order.CancelPreOrderResp{}, nil
}
