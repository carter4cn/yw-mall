package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
	// S1 cross-DB read of mall_order for the cashier.
	OrderDataSource string `json:",optional"`
	// S1.8 feature flag — set false in production.
	PaymentMockEnabled bool   `json:",default=true"`
	DefaultChannel     string `json:",default=mock"`
}
