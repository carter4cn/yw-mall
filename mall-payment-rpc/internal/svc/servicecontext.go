package svc

import (
	"mall-payment-rpc/internal/channel"
	"mall-payment-rpc/internal/config"
	"mall-payment-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	PaymentModel model.PaymentModel
	SqlConn      sqlx.SqlConn
	OrderDB      sqlx.SqlConn // S1: cross-DB read of mall_order
	Channels     map[string]channel.PayChannel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	ctx := &ServiceContext{
		Config:       c,
		PaymentModel: model.NewPaymentModel(conn, c.Cache),
		SqlConn:      conn,
		Channels:     map[string]channel.PayChannel{"mock": &channel.MockChannel{}},
	}
	if c.OrderDataSource != "" {
		ctx.OrderDB = sqlx.NewMysql(c.OrderDataSource)
	}
	return ctx
}
