package logic

import (
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"
)

func protoShipmentToType(s *logisticspb.Shipment) types.ShipmentDTO {
	out := types.ShipmentDTO{
		Id:              s.Id,
		OrderId:         s.OrderId,
		UserId:          s.UserId,
		TrackingNo:      s.TrackingNo,
		Carrier:         s.Carrier,
		Status:          s.Status,
		SubscribeStatus: s.SubscribeStatus,
		LastTrackTime:   s.LastTrackTime,
		CreateTime:      s.CreateTime,
	}
	for _, it := range s.Items {
		out.Items = append(out.Items, types.ShipmentItemRef{
			OrderItemId: it.OrderItemId,
			ProductId:   it.ProductId,
			Quantity:    it.Quantity,
		})
	}
	for _, t := range s.Tracks {
		out.Tracks = append(out.Tracks, types.ShipmentTrack{
			TrackTime:      t.TrackTime,
			Location:       t.Location,
			Description:    t.Description,
			StateInternal:  t.StateInternal,
			StateKuaidi100: t.StateKuaidi100,
		})
	}
	return out
}
