package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShopModel = (*customShopModel)(nil)

type (
	// ShopModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShopModel.
	ShopModel interface {
		shopModel
	}

	customShopModel struct {
		*defaultShopModel
	}
)

// NewShopModel returns a model for the database table.
func NewShopModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ShopModel {
	return &customShopModel{
		defaultShopModel: newShopModel(conn, c, opts...),
	}
}
