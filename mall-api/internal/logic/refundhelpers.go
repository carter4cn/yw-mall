package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/types"
	"mall-order-rpc/order"
)

// uidFromContext lifts the authenticated uid (set by SessionAuth middleware)
// into an int64. Returns 0 if absent so logic layers can decide whether to
// reject. P0 login revamp: replaces the old JWT-claim path.
func uidFromContext(ctx context.Context) int64 {
	return middleware.UidFromCtx(ctx)
}

// refundProtoToDTO converts the protobuf RefundRequest into the JSON wire type
// the C-side frontend consumes.
func refundProtoToDTO(r *order.RefundRequest) types.RefundRequestDTO {
	if r == nil {
		return types.RefundRequestDTO{}
	}
	items := make([]types.RefundItemDTO, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, types.RefundItemDTO{
			SkuId:    it.SkuId,
			SkuName:  it.SkuName,
			Quantity: it.Quantity,
			Amount:   it.Amount,
		})
	}
	return types.RefundRequestDTO{
		Id:                 r.Id,
		OrderId:            r.OrderId,
		OrderNo:            r.OrderNo,
		UserId:             r.UserId,
		ShopId:             r.ShopId,
		Amount:             r.Amount,
		Reason:             r.Reason,
		Evidence:           append([]string{}, r.Evidence...),
		Items:              items,
		Status:             r.Status,
		MerchantRemark:     r.MerchantRemark,
		MerchantHandleTime: r.MerchantHandleTime,
		AdminRemark:        r.AdminRemark,
		AdminHandleTime:    r.AdminHandleTime,
		AppealReason:       r.AppealReason,
		AppealTime:         r.AppealTime,
		RefundNo:           r.RefundNo,
		RefundCompleteTime: r.RefundCompleteTime,
		CreateTime:         r.CreateTime,
	}
}
