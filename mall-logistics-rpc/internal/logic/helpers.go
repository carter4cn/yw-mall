package logic

import (
	"errors"
	"mall-logistics-rpc/internal/model"
	"mall-logistics-rpc/logistics"

	gosqldriver "github.com/go-sql-driver/mysql"
)

func isDuplicateKey(err error) bool {
	var me *gosqldriver.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return true
	}
	return false
}

// toShipmentProto converts a model.Shipment plus its items and tracks into the proto Shipment.
func toShipmentProto(s *model.Shipment, items []*model.ShipmentItem, tracks []*model.ShipmentTrack) *logistics.Shipment {
	out := &logistics.Shipment{
		Id:              s.Id,
		OrderId:         s.OrderId,
		UserId:          s.UserId,
		TrackingNo:      s.TrackingNo,
		Carrier:         s.Carrier,
		Status:          int32(s.Status),
		SubscribeStatus: int32(s.SubscribeStatus),
		CreateTime:      s.CreateTime.Unix(),
	}
	if s.LastTrackTime.Valid {
		out.LastTrackTime = s.LastTrackTime.Time.Unix()
	}
	for _, it := range items {
		out.Items = append(out.Items, &logistics.ShipmentItemRef{
			OrderItemId: it.OrderItemId,
			ProductId:   it.ProductId,
			Quantity:    int32(it.Quantity),
		})
	}
	for _, t := range tracks {
		var k100 int32
		if t.StateKuaidi100.Valid {
			k100 = int32(t.StateKuaidi100.Int64)
		}
		var loc string
		if t.Location.Valid {
			loc = t.Location.String
		}
		out.Tracks = append(out.Tracks, &logistics.Track{
			TrackTime:      t.TrackTime.Unix(),
			Location:       loc,
			Description:    t.Description,
			StateInternal:  int32(t.StateInternal),
			StateKuaidi100: k100,
		})
	}
	return out
}
