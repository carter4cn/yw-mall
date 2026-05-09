package svc

import (
	"mall-shop-rpc/internal/config"
	"mall-shop-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	ShopModel       model.ShopModel
	ShopFollowModel model.ShopFollowModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:          c,
		DB:              conn,
		ShopModel:       model.NewShopModel(conn, c.Cache),
		ShopFollowModel: model.NewShopFollowModel(conn, c.Cache),
	}
}
