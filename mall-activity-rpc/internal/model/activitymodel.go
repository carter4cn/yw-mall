package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ActivityModel = (*customActivityModel)(nil)

type (
	// ActivityModel is an interface to be customized, add more methods here,
	// and implement the added methods in customActivityModel.
	ActivityModel interface {
		activityModel
	}

	customActivityModel struct {
		*defaultActivityModel
	}
)

// NewActivityModel returns a model for the database table.
func NewActivityModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ActivityModel {
	return &customActivityModel{
		defaultActivityModel: newActivityModel(conn, c, opts...),
	}
}
