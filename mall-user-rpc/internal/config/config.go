package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
	JwtAuth    struct {
		AccessSecret string
	}
	// Session: Redis backing store for opaque-token sessions (P0 login revamp).
	// Reuses the same Redis cluster as `Cache` but with separate logical keys.
	Session struct {
		Redis struct {
			Host string
			Pass string
			DB   int
		}
		AccessTTLSeconds  int64 // default 1800 (30 min)
		RefreshTTLSeconds int64 // default 604800 (7 days)
		MaxRotateCount    int32 // default 10
	}
}
