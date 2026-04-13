// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.OrderDetailReq) (resp *types.OrderDetailResp, err error) {
	res, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &order.GetOrderReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.CreateOrderItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, types.CreateOrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	return &types.OrderDetailResp{
		Id:          res.Id,
		OrderNo:     res.OrderNo,
		UserId:      res.UserId,
		TotalAmount: res.TotalAmount,
		Status:      res.Status,
		Items:       items,
		CreateTime:  res.CreateTime,
	}, nil
}
