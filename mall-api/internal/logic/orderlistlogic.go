// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderListLogic) OrderList(req *types.OrderListReq) (resp *types.OrderListResp, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.OrderRpc.ListOrders(l.ctx, &order.ListOrdersReq{
		UserId:   userId,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	orders := make([]types.OrderDetailResp, 0, len(res.Orders))
	for _, o := range res.Orders {
		items := make([]types.CreateOrderItem, 0, len(o.Items))
		for _, item := range o.Items {
			items = append(items, types.CreateOrderItem{
				ProductId:   item.ProductId,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}
		orders = append(orders, types.OrderDetailResp{
			Id:          o.Id,
			OrderNo:     o.OrderNo,
			UserId:      o.UserId,
			TotalAmount: o.TotalAmount,
			Status:      o.Status,
			Items:       items,
			CreateTime:  o.CreateTime,
		})
	}

	return &types.OrderListResp{
		Orders: orders,
		Total:  res.Total,
	}, nil
}
