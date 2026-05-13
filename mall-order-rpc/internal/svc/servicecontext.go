package svc

import (
	"mall-order-rpc/internal/config"
	"mall-order-rpc/internal/kafka"
	"mall-order-rpc/internal/model"
	"mall-payment-rpc/paymentclient"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config               config.Config
	OrderModel           model.OrderModel
	OrderItemModel       model.OrderItemModel
	SqlConn              sqlx.SqlConn
	OrderShippedProducer *kafka.Producer
	UserRpc              userclient.User
	// S2: optional — refund flows debit merchant wallet via mall-payment-rpc.
	PaymentRpc paymentclient.Payment
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	ctx := &ServiceContext{
		Config:               c,
		OrderModel:           model.NewOrderModel(conn, c.Cache),
		OrderItemModel:       model.NewOrderItemModel(conn, c.Cache),
		SqlConn:              conn,
		OrderShippedProducer: kafka.NewProducer(c.Kafka.Brokers, c.Kafka.OrderShippedTopic),
		UserRpc:              userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
	if c.PaymentRpc.Etcd.Key != "" || len(c.PaymentRpc.Endpoints) > 0 {
		ctx.PaymentRpc = paymentclient.NewPayment(zrpc.MustNewClient(c.PaymentRpc))
	}
	return ctx
}
