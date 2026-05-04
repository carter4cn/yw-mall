package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	orderStatusPaid    = 1
	orderStatusShipped = 2
)

type orderShippedItem struct {
	OrderItemId int64 `json:"orderItemId"`
	ProductId   int64 `json:"productId"`
	Quantity    int32 `json:"quantity"`
}

type orderShippedEvent struct {
	OrderId    int64              `json:"orderId"`
	UserId     int64              `json:"userId"`
	TrackingNo string             `json:"trackingNo"`
	Carrier    string             `json:"carrier"`
	Items      []orderShippedItem `json:"items"`
	ShippedAt  int64              `json:"shippedAt"`
}

type MarkShippedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkShippedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkShippedLogic {
	return &MarkShippedLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *MarkShippedLogic) MarkShipped(in *order.MarkShippedReq) (*order.MarkShippedResp, error) {
	if strings.TrimSpace(in.TrackingNo) == "" || strings.TrimSpace(in.Carrier) == "" {
		return nil, errors.New("tracking_no and carrier are required")
	}
	ord, err := l.svcCtx.OrderModel.FindOne(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, errors.New("order not found")
	}
	if int(ord.Status) != orderStatusPaid {
		return nil, errors.New("order not in a shippable state")
	}
	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE `order` SET status=?, tracking_no=?, carrier=? WHERE id=? AND status=?",
		orderStatusShipped, in.TrackingNo, in.Carrier, in.OrderId, orderStatusPaid); err != nil {
		return nil, err
	}
	items, err := l.fetchItems(in.OrderId)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(orderShippedEvent{
		OrderId:    in.OrderId,
		UserId:     int64(ord.UserId),
		TrackingNo: in.TrackingNo,
		Carrier:    in.Carrier,
		Items:      items,
		ShippedAt:  time.Now().Unix(),
	})
	if err := l.svcCtx.OrderShippedProducer.Write(l.ctx, in.TrackingNo, body); err != nil {
		return nil, err
	}
	return &order.MarkShippedResp{Ok: true}, nil
}

func (l *MarkShippedLogic) fetchItems(orderId int64) ([]orderShippedItem, error) {
	rows := []*struct {
		Id        int64 `db:"id"`
		ProductId int64 `db:"product_id"`
		Quantity  int32 `db:"quantity"`
	}{}
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, product_id, quantity FROM order_item WHERE order_id=?", orderId); err != nil {
		return nil, err
	}
	out := make([]orderShippedItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, orderShippedItem{OrderItemId: r.Id, ProductId: r.ProductId, Quantity: r.Quantity})
	}
	return out, nil
}
