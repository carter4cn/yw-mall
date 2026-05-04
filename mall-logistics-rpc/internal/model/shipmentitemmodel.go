package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShipmentItemModel = (*customShipmentItemModel)(nil)

type (
	// ShipmentItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShipmentItemModel.
	ShipmentItemModel interface {
		shipmentItemModel
	}

	customShipmentItemModel struct {
		*defaultShipmentItemModel
	}
)

// NewShipmentItemModel returns a model for the database table.
func NewShipmentItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ShipmentItemModel {
	return &customShipmentItemModel{
		defaultShipmentItemModel: newShipmentItemModel(conn, c, opts...),
	}
}
