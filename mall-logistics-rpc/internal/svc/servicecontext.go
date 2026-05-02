package svc

import (
	"net/http"
	"time"

	"mall-logistics-rpc/internal/config"
	"mall-logistics-rpc/internal/kuaidi100"
	"mall-logistics-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	DB                 sqlx.SqlConn
	ShipmentModel      model.ShipmentModel
	ShipmentItemModel  model.ShipmentItemModel
	ShipmentTrackModel model.ShipmentTrackModel
	Redis              *redis.Redis
	Kuaidi100          *kuaidi100.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:             c,
		DB:                 conn,
		ShipmentModel:      model.NewShipmentModel(conn, c.Cache),
		ShipmentItemModel:  model.NewShipmentItemModel(conn, c.Cache),
		ShipmentTrackModel: model.NewShipmentTrackModel(conn, c.Cache),
		Redis:              redis.MustNewRedis(c.RedisCache),
		Kuaidi100: kuaidi100.NewClient(kuaidi100.Config{
			Customer:        c.Kuaidi100.Customer,
			Key:             c.Kuaidi100.Key,
			PollEndpoint:    c.Kuaidi100.PollEndpoint,
			WebhookCallback: c.Kuaidi100.WebhookCallback,
			HTTP:            &http.Client{Timeout: 10 * time.Second},
		}),
	}
}
