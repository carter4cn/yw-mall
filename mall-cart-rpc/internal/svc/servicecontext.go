package svc

import (
	"mall-cart-rpc/internal/config"
	"mall-cart-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	CartItemModel model.CartItemModel
	SqlConn       sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:        c,
		CartItemModel: model.NewCartItemModel(conn, c.Cache),
		SqlConn:       conn,
	}
}
