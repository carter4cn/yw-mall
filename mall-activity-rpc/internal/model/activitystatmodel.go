package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ActivityStatModel = (*customActivityStatModel)(nil)

type (
	// ActivityStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customActivityStatModel.
	ActivityStatModel interface {
		activityStatModel
	}

	customActivityStatModel struct {
		*defaultActivityStatModel
	}
)

// NewActivityStatModel returns a model for the database table.
func NewActivityStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ActivityStatModel {
	return &customActivityStatModel{
		defaultActivityStatModel: newActivityStatModel(conn, c, opts...),
	}
}
