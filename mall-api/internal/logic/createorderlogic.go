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

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	items := make([]*order.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &order.OrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	res, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &order.CreateOrderReq{
		UserId: userId,
		Items:  items,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateOrderResp{
		Id:          res.Id,
		OrderNo:     res.OrderNo,
		TotalAmount: res.TotalAmount,
	}, nil
}
