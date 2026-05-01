package svc

import (
	"mall-risk-rpc/internal/config"
	"mall-risk-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                  config.Config
	DB                      sqlx.SqlConn
	Redis                   *redis.Redis
	BlacklistModel          model.BlacklistModel
	RateLimitConfigModel    model.RateLimitConfigModel
	RiskScoreCacheModel     model.RiskScoreCacheModel
	ParticipationTokenModel model.ParticipationTokenModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rds := redis.MustNewRedis(c.RedisCache)
	return &ServiceContext{
		Config:                  c,
		DB:                      conn,
		Redis:                   rds,
		BlacklistModel:          model.NewBlacklistModel(conn, c.Cache),
		RateLimitConfigModel:    model.NewRateLimitConfigModel(conn, c.Cache),
		RiskScoreCacheModel:     model.NewRiskScoreCacheModel(conn, c.Cache),
		ParticipationTokenModel: model.NewParticipationTokenModel(conn, c.Cache),
	}
}
