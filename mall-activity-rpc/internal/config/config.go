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
	WorkflowRpc zrpc.RpcClientConf
	RiskRpc     zrpc.RpcClientConf
	RewardRpc   zrpc.RpcClientConf
	// RewardTemplates maps activity_type -> template code; the participate
	// path looks the template id up by code at startup so seed-binary order
	// can't poison this mapping with stale ids.
	RewardTemplates struct {
		Signin  string
		Lottery string
		Coupon  string
	}
}
