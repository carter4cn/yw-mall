package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShipmentModel = (*customShipmentModel)(nil)

type (
	// ShipmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShipmentModel.
	ShipmentModel interface {
		shipmentModel
	}

	customShipmentModel struct {
		*defaultShipmentModel
	}
)

// NewShipmentModel returns a model for the database table.
func NewShipmentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ShipmentModel {
	return &customShipmentModel{
		defaultShipmentModel: newShipmentModel(conn, c, opts...),
	}
}
