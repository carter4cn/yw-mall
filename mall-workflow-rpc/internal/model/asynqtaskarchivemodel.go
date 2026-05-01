package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AsynqTaskArchiveModel = (*customAsynqTaskArchiveModel)(nil)

type (
	// AsynqTaskArchiveModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAsynqTaskArchiveModel.
	AsynqTaskArchiveModel interface {
		asynqTaskArchiveModel
	}

	customAsynqTaskArchiveModel struct {
		*defaultAsynqTaskArchiveModel
	}
)

// NewAsynqTaskArchiveModel returns a model for the database table.
func NewAsynqTaskArchiveModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AsynqTaskArchiveModel {
	return &customAsynqTaskArchiveModel{
		defaultAsynqTaskArchiveModel: newAsynqTaskArchiveModel(conn, c, opts...),
	}
}
