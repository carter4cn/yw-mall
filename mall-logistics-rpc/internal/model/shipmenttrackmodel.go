package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShipmentTrackModel = (*customShipmentTrackModel)(nil)

type (
	// ShipmentTrackModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShipmentTrackModel.
	ShipmentTrackModel interface {
		shipmentTrackModel
	}

	customShipmentTrackModel struct {
		*defaultShipmentTrackModel
	}
)

// NewShipmentTrackModel returns a model for the database table.
func NewShipmentTrackModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ShipmentTrackModel {
	return &customShipmentTrackModel{
		defaultShipmentTrackModel: newShipmentTrackModel(conn, c, opts...),
	}
}
