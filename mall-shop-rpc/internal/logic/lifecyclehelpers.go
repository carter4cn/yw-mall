package logic

import "mall-shop-rpc/shop"

type lifecycleRequestRow struct {
	Id          uint64 `db:"id"`
	ShopId      int64  `db:"shop_id"`
	Action      string `db:"action"`
	Reason      string `db:"reason"`
	Status      int64  `db:"status"`
	AdminId     int64  `db:"admin_id"`
	AdminRemark string `db:"admin_remark"`
	CreateTime  int64  `db:"create_time"`
	UpdateTime  int64  `db:"update_time"`
}

const lifecycleRequestCols = "id, shop_id, action, reason, status, admin_id, admin_remark, create_time, update_time"

func toLifecycleRequestProto(r *lifecycleRequestRow) *shop.ShopLifecycleRequest {
	return &shop.ShopLifecycleRequest{
		Id:          int64(r.Id),
		ShopId:      r.ShopId,
		Action:      r.Action,
		Reason:      r.Reason,
		Status:      int32(r.Status),
		AdminId:     r.AdminId,
		AdminRemark: r.AdminRemark,
		CreateTime:  r.CreateTime,
	}
}
