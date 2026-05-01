package svc

import (
	"mall-user-rpc/internal/config"
	"mall-user-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	DB        sqlx.SqlConn
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:    c,
		DB:        conn,
		UserModel: model.NewUserModel(conn, c.Cache),
	}
}
