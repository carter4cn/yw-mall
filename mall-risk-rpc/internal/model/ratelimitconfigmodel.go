package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RateLimitConfigModel = (*customRateLimitConfigModel)(nil)

type (
	// RateLimitConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRateLimitConfigModel.
	RateLimitConfigModel interface {
		rateLimitConfigModel
	}

	customRateLimitConfigModel struct {
		*defaultRateLimitConfigModel
	}
)

// NewRateLimitConfigModel returns a model for the database table.
func NewRateLimitConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RateLimitConfigModel {
	return &customRateLimitConfigModel{
		defaultRateLimitConfigModel: newRateLimitConfigModel(conn, c, opts...),
	}
}
