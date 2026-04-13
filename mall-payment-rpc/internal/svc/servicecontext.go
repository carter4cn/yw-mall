package svc

import (
	"mall-payment-rpc/internal/config"
	"mall-payment-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	PaymentModel model.PaymentModel
	SqlConn      sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:       c,
		PaymentModel: model.NewPaymentModel(conn, c.Cache),
		SqlConn:      conn,
	}
}
