package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ActivityInventorySnapshotModel = (*customActivityInventorySnapshotModel)(nil)

type (
	// ActivityInventorySnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customActivityInventorySnapshotModel.
	ActivityInventorySnapshotModel interface {
		activityInventorySnapshotModel
	}

	customActivityInventorySnapshotModel struct {
		*defaultActivityInventorySnapshotModel
	}
)

// NewActivityInventorySnapshotModel returns a model for the database table.
func NewActivityInventorySnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ActivityInventorySnapshotModel {
	return &customActivityInventorySnapshotModel{
		defaultActivityInventorySnapshotModel: newActivityInventorySnapshotModel(conn, c, opts...),
	}
}
