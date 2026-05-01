package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BlacklistModel = (*customBlacklistModel)(nil)

type (
	// BlacklistModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlacklistModel.
	BlacklistModel interface {
		blacklistModel
	}

	customBlacklistModel struct {
		*defaultBlacklistModel
	}
)

// NewBlacklistModel returns a model for the database table.
func NewBlacklistModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BlacklistModel {
	return &customBlacklistModel{
		defaultBlacklistModel: newBlacklistModel(conn, c, opts...),
	}
}
