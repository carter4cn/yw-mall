package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/internal/util"
	"mall-order-rpc/order"

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
	orderNo := util.GenerateOrderNo()

	var totalAmount int64
	for _, item := range in.Items {
		totalAmount += item.Price * int64(item.Quantity)
	}

	var orderId int64
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			"INSERT INTO `order` (`order_no`, `user_id`, `total_amount`, `status`) VALUES (?, ?, ?, ?)",
			orderNo, in.UserId, totalAmount, 0,
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
