package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	DataSource string
	Cache      cache.CacheConf
	RedisCache redis.RedisConf

	OrderRpc zrpc.RpcClientConf
	RiskRpc  zrpc.RpcClientConf

	Followup struct {
		MinDelayDays int
		MaxLength    int
	}
	Review struct {
		ContentMin int
		ContentMax int
	}

	CacheTTLSeconds int
	MediaUrlPrefix  string
}
