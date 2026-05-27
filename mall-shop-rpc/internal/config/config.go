package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource     string
	RiskDataSource string `json:",optional"`
	UserDataSource string `json:",optional"` // mall_user 库（merchant_staff 表）
	Cache          cache.CacheConf
}
