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

	Kafka struct {
		Brokers []string
		Topic   string
		Group   string
	}
	Kuaidi100 struct {
		Customer        string
		Key             string
		PollEndpoint    string
		WebhookCallback string
	}
	Subscribe struct {
		MaxRetries       int
		InitialBackoffMs int
	}
}
