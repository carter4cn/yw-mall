package svc

import (
	"mall-shop-rpc/internal/config"
	"mall-shop-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	RiskDB          sqlx.SqlConn // for F-5 auto-restrict via shop_restriction
	UserDB          sqlx.SqlConn // M1: mall_user 库（merchant_staff + merchant_staff_invitation）
	ShopModel       model.ShopModel
	ShopFollowModel model.ShopFollowModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	ctx := &ServiceContext{
		Config:          c,
		DB:              conn,
		ShopModel:       model.NewShopModel(conn, c.Cache),
		ShopFollowModel: model.NewShopFollowModel(conn, c.Cache),
	}
	if c.RiskDataSource != "" {
		ctx.RiskDB = sqlx.NewMysql(c.RiskDataSource)
	}
	if c.UserDataSource != "" {
		ctx.UserDB = sqlx.NewMysql(c.UserDataSource)
	}
	return ctx
}
