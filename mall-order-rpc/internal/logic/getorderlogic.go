package logic

import (
	"context"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *order.GetOrderReq) (*order.GetOrderResp, error) {
	o, err := l.svcCtx.OrderModel.FindOne(l.ctx, uint64(in.Id))
	if err == model.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var items []*model.OrderItem
	err = l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &items,
		"SELECT `id`, `order_id`, `product_id`, `product_name`, `price`, `quantity`, `create_time` FROM `order_item` WHERE `order_id` = ?",
		o.Id,
	)
	if err != nil {
		return nil, err
	}

	pbItems := make([]*order.OrderItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &order.OrderItem{
			ProductId:   int64(item.ProductId),
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    int32(item.Quantity),
		})
	}

	return &order.GetOrderResp{
		Id:          int64(o.Id),
		OrderNo:     o.OrderNo,
		UserId:      int64(o.UserId),
		TotalAmount: o.TotalAmount,
		Status:      int32(o.Status),
		Items:       pbItems,
		CreateTime:  o.CreateTime.Unix(),
	}, nil
}
