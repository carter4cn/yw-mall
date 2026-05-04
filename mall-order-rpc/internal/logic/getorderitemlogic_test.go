package logic

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"
)

// mockOrderItemModel is a minimal stub for OrderItemModel.
type mockOrderItemModel struct {
	item *model.OrderItem
	err  error
}

func (m *mockOrderItemModel) Insert(ctx context.Context, data *model.OrderItem) (sql.Result, error) {
	return nil, nil
}
func (m *mockOrderItemModel) FindOne(ctx context.Context, id uint64) (*model.OrderItem, error) {
	return m.item, m.err
}
func (m *mockOrderItemModel) Update(ctx context.Context, data *model.OrderItem) error { return nil }
func (m *mockOrderItemModel) Delete(ctx context.Context, id uint64) error             { return nil }

// mockOrderModel is a minimal stub for OrderModel.
type mockOrderModel struct {
	ord *model.Order
	err error
}

func (m *mockOrderModel) Insert(ctx context.Context, data *model.Order) (sql.Result, error) {
	return nil, nil
}
func (m *mockOrderModel) FindOne(ctx context.Context, id uint64) (*model.Order, error) {
	return m.ord, m.err
}
func (m *mockOrderModel) FindOneByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	return nil, nil
}
func (m *mockOrderModel) Update(ctx context.Context, data *model.Order) error { return nil }
func (m *mockOrderModel) Delete(ctx context.Context, id uint64) error         { return nil }

func TestGetOrderItem_ItemNotFound(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		OrderItemModel: &mockOrderItemModel{err: model.ErrNotFound},
		OrderModel:     &mockOrderModel{},
	}
	l := NewGetOrderItemLogic(context.Background(), svcCtx)
	_, err := l.GetOrderItem(&order.GetOrderItemReq{OrderItemId: 999})
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestGetOrderItem_OrderNotFound(t *testing.T) {
	item := &model.OrderItem{Id: 1, OrderId: 42, ProductId: 7, Quantity: 2}
	svcCtx := &svc.ServiceContext{
		OrderItemModel: &mockOrderItemModel{item: item},
		OrderModel:     &mockOrderModel{err: model.ErrNotFound},
	}
	l := NewGetOrderItemLogic(context.Background(), svcCtx)
	_, err := l.GetOrderItem(&order.GetOrderItemReq{OrderItemId: 1})
	assert.ErrorIs(t, err, model.ErrNotFound)
}
