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
	Kafka      struct {
		Brokers []string
	}
	Dtm struct {
		Server string
	}
	RuleRpc     zrpc.RpcClientConf
	RewardRpc   zrpc.RpcClientConf
	ActivityRpc zrpc.RpcClientConf
	UserRpc     zrpc.RpcClientConf
	ProductRpc  zrpc.RpcClientConf
	OrderRpc    zrpc.RpcClientConf
}
