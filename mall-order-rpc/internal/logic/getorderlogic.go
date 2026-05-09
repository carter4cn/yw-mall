package logic

import (
	"context"
	"time"

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

type orderRow struct {
	Id               uint64    `db:"id"`
	OrderNo          string    `db:"order_no"`
	UserId           uint64    `db:"user_id"`
	TotalAmount      int64     `db:"total_amount"`
	Status           int64     `db:"status"`
	CreateTime       time.Time `db:"create_time"`
	AddressId        int64     `db:"address_id"`
	ReceiverName     string    `db:"receiver_name"`
	ReceiverPhone    string    `db:"receiver_phone"`
	ReceiverProvince string    `db:"receiver_province"`
	ReceiverCity     string    `db:"receiver_city"`
	ReceiverDistrict string    `db:"receiver_district"`
	ReceiverDetail   string    `db:"receiver_detail"`
}

func (l *GetOrderLogic) GetOrder(in *order.GetOrderReq) (*order.GetOrderResp, error) {
	var o orderRow
	err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &o,
		"SELECT `id`, `order_no`, `user_id`, `total_amount`, `status`, `create_time`, `address_id`, `receiver_name`, `receiver_phone`, `receiver_province`, `receiver_city`, `receiver_district`, `receiver_detail` FROM `order` WHERE `id` = ? LIMIT 1",
		in.Id,
	)
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
		Id:               int64(o.Id),
		OrderNo:          o.OrderNo,
		UserId:           int64(o.UserId),
		TotalAmount:      o.TotalAmount,
		Status:           int32(o.Status),
		Items:            pbItems,
		CreateTime:       o.CreateTime.Unix(),
		AddressId:        o.AddressId,
		ReceiverName:     o.ReceiverName,
		ReceiverPhone:    o.ReceiverPhone,
		ReceiverProvince: o.ReceiverProvince,
		ReceiverCity:     o.ReceiverCity,
		ReceiverDistrict: o.ReceiverDistrict,
		ReceiverDetail:   o.ReceiverDetail,
	}, nil
}
