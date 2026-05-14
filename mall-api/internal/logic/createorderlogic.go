// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"errors"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"
	"mall-user-rpc/userclient"

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
	userId := middleware.UidFromCtx(l.ctx)

	// Pick the user's default address. order-rpc requires a concrete address_id
	// and snapshots receiver fields onto the order row; without this the call
	// fails with "address not found" because order-rpc gets address_id=0.
	addr, err := l.svcCtx.UserRpc.GetDefaultAddress(l.ctx, &userclient.GetDefaultAddressReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	if addr == nil || addr.Id == 0 {
		return nil, errors.New("请先添加默认收货地址")
	}

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
		UserId:    userId,
		AddressId: addr.Id,
		Items:     items,
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
