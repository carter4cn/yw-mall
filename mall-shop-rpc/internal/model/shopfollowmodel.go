package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShopFollowModel = (*customShopFollowModel)(nil)

type (
	// ShopFollowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShopFollowModel.
	ShopFollowModel interface {
		shopFollowModel
	}

	customShopFollowModel struct {
		*defaultShopFollowModel
	}
)

// NewShopFollowModel returns a model for the database table.
func NewShopFollowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ShopFollowModel {
	return &customShopFollowModel{
		defaultShopFollowModel: newShopFollowModel(conn, c, opts...),
	}
}
