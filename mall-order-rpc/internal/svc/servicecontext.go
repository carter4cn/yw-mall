package svc

import (
	"mall-order-rpc/internal/config"
	"mall-order-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	OrderModel     model.OrderModel
	OrderItemModel model.OrderItemModel
	SqlConn        sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:         c,
		OrderModel:     model.NewOrderModel(conn, c.Cache),
		OrderItemModel: model.NewOrderItemModel(conn, c.Cache),
		SqlConn:        conn,
	}
}
