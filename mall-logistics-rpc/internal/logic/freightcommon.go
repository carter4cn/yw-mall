package logic

import "mall-logistics-rpc/logistics"

type freightTemplateRow struct {
	Id         uint64 `db:"id"`
	ShopId     int64  `db:"shop_id"`
	Name       string `db:"name"`
	CalcType   int64  `db:"calc_type"`
	FirstValue int64  `db:"first_value"`
	FirstFee   int64  `db:"first_fee"`
	ExtraValue int64  `db:"extra_value"`
	ExtraFee   int64  `db:"extra_fee"`
	Regions    string `db:"regions"`
	IsDefault  int64  `db:"is_default"`
	Status     int64  `db:"status"`
	CreateTime int64  `db:"create_time"`
	UpdateTime int64  `db:"update_time"`
}

const freightTemplateCols = "id, shop_id, name, calc_type, first_value, first_fee, extra_value, extra_fee, regions, is_default, status, create_time, update_time"

func toFreightTemplateProto(r *freightTemplateRow) *logistics.FreightTemplate {
	return &logistics.FreightTemplate{
		Id:         int64(r.Id),
		ShopId:     r.ShopId,
		Name:       r.Name,
		CalcType:   int32(r.CalcType),
		FirstValue: int32(r.FirstValue),
		FirstFee:   r.FirstFee,
		ExtraValue: int32(r.ExtraValue),
		ExtraFee:   r.ExtraFee,
		Regions:    r.Regions,
		IsDefault:  r.IsDefault == 1,
		Status:     int32(r.Status),
		CreateTime: r.CreateTime,
	}
}

func normPage(p, s int32) (int32, int32) {
	if p <= 0 {
		p = 1
	}
	if s <= 0 || s > 50 {
		s = 20
	}
	return p, s
}
