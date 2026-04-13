package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CartItemModel = (*customCartItemModel)(nil)

type (
	// CartItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCartItemModel.
	CartItemModel interface {
		cartItemModel
	}

	customCartItemModel struct {
		*defaultCartItemModel
	}
)

// NewCartItemModel returns a model for the database table.
func NewCartItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CartItemModel {
	return &customCartItemModel{
		defaultCartItemModel: newCartItemModel(conn, c, opts...),
	}
}
