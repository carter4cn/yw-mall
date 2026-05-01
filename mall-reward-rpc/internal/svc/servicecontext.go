package svc

import (
	"mall-order-rpc/orderclient"
	"mall-product-rpc/productclient"
	"mall-reward-rpc/internal/config"
	"mall-reward-rpc/internal/model"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                 config.Config
	DB                     sqlx.SqlConn
	Redis                  *redis.Redis
	RewardTemplateModel    model.RewardTemplateModel
	RewardRecordModel      model.RewardRecordModel
	RewardDispatchLogModel model.RewardDispatchLogModel
	OutboxModel            model.OutboxModel
	UserRpc                userclient.User
	ProductRpc             productclient.Product
	OrderRpc               orderclient.Order
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rds := redis.MustNewRedis(c.RedisCache)
	return &ServiceContext{
		Config:                 c,
		DB:                     conn,
		Redis:                  rds,
		RewardTemplateModel:    model.NewRewardTemplateModel(conn, c.Cache),
		RewardRecordModel:      model.NewRewardRecordModel(conn, c.Cache),
		RewardDispatchLogModel: model.NewRewardDispatchLogModel(conn, c.Cache),
		OutboxModel:            model.NewOutboxModel(conn, c.Cache),
		UserRpc:                userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc:             productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		OrderRpc:               orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
	}
}
