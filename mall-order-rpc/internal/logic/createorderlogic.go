package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/internal/util"
	"mall-order-rpc/order"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	addr, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &userclient.GetAddressReq{
		UserId: in.UserId,
		Id:     in.AddressId,
	})
	if err != nil {
		return nil, err
	}

	orderNo := util.GenerateOrderNo()

	var totalAmount int64
	for _, item := range in.Items {
		totalAmount += item.Price * int64(item.Quantity)
	}

	var orderId int64
	err = l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			"INSERT INTO `order` (`order_no`, `user_id`, `total_amount`, `status`, `address_id`, `receiver_name`, `receiver_phone`, `receiver_province`, `receiver_city`, `receiver_district`, `receiver_detail`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			orderNo, in.UserId, totalAmount, 0,
			addr.Id, addr.ReceiverName, addr.Phone,
			addr.Province, addr.City, addr.District, addr.Detail,
		)
		if err != nil {
			return err
		}

		orderId, err = result.LastInsertId()
		if err != nil {
			return err
		}

		for _, item := range in.Items {
			_, err = session.ExecCtx(ctx,
				"INSERT INTO `order_item` (`order_id`, `product_id`, `product_name`, `price`, `quantity`) VALUES (?, ?, ?, ?, ?)",
				orderId, item.ProductId, item.ProductName, item.Price, item.Quantity,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order.CreateOrderResp{
		Id:          orderId,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
	}, nil
}
