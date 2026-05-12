package logic

import (
	"context"
	"errors"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetShopOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopOrderLogic {
	return &GetShopOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetShopOrderLogic) GetShopOrder(in *order.GetShopOrderReq) (*order.GetOrderResp, error) {
	var o orderRow
	err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &o,
		"SELECT "+orderTimelineCols+" FROM `order` WHERE `id` = ? AND `shop_id` = ? LIMIT 1",
		in.Id, in.ShopId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("order not found or not owned by shop")
		}
		return nil, err
	}

	var items []*model.OrderItem
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &items,
		"SELECT `id`, `order_id`, `product_id`, `product_name`, `price`, `quantity`, `create_time` FROM `order_item` WHERE `order_id` = ?",
		o.Id); err != nil {
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
		PayTime:          o.PayTime,
		ShipTime:         o.ShipTime,
		CompleteTime:     o.CompleteTime,
		CancelTime:       o.CancelTime,
		CancelReason:     o.CancelReason,
	}, nil
}
