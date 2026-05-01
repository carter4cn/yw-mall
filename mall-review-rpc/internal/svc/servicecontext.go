package svc

import (
	"mall-order-rpc/orderclient"
	"mall-review-rpc/internal/config"
	"mall-review-rpc/internal/model"
	"mall-risk-rpc/riskclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	ReviewModel      model.ReviewModel
	ReviewMediaModel model.ReviewMediaModel

	OrderRpc orderclient.Order
	RiskRpc  riskclient.Risk

	Redis *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:           c,
		ReviewModel:      model.NewReviewModel(conn, c.Cache),
		ReviewMediaModel: model.NewReviewMediaModel(conn, c.Cache),
		OrderRpc:         orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		RiskRpc:          riskclient.NewRisk(zrpc.MustNewClient(c.RiskRpc)),
		Redis:            redis.MustNewRedis(c.RedisCache),
	}
}
