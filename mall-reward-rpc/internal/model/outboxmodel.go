package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OutboxModel = (*customOutboxModel)(nil)

type (
	// OutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOutboxModel.
	OutboxModel interface {
		outboxModel
	}

	customOutboxModel struct {
		*defaultOutboxModel
	}
)

// NewOutboxModel returns a model for the database table.
func NewOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) OutboxModel {
	return &customOutboxModel{
		defaultOutboxModel: newOutboxModel(conn, c, opts...),
	}
}
