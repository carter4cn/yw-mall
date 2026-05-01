package logic

import (
	"context"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderItemLogic {
	return &GetOrderItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderItemLogic) GetOrderItem(in *order.GetOrderItemReq) (*order.GetOrderItemResp, error) {
	item, err := l.svcCtx.OrderItemModel.FindOne(l.ctx, uint64(in.OrderItemId))
	if err == model.ErrNotFound {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	ord, err := l.svcCtx.OrderModel.FindOne(l.ctx, item.OrderId)
	if err == model.ErrNotFound {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &order.GetOrderItemResp{
		OrderItemId: int64(item.Id),
		OrderId:     int64(item.OrderId),
		UserId:      int64(ord.UserId),
		ProductId:   int64(item.ProductId),
		Quantity:    item.Quantity,
		OrderStatus: int32(ord.Status),
		CreateTime:  ord.CreateTime.Unix(),
	}, nil
}
