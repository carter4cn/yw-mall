package logic

import (
	"mall-shop-rpc/shop"
)

type levelTemplateRow struct {
	Id             uint64  `db:"id"`
	Level          int64   `db:"level"`
	Name           string  `db:"name"`
	MinGmv         int64   `db:"min_gmv"`
	MinCreditScore int64   `db:"min_credit_score"`
	MinMonths      int64   `db:"min_months"`
	MinRating      float64 `db:"min_rating"`
	CommissionRate float64 `db:"commission_rate"`
	TrafficBoost   float64 `db:"traffic_boost"`
	Benefits       string  `db:"benefits"`
	CreateTime     int64   `db:"create_time"`
}

const levelTemplateCols = "id, level, name, min_gmv, min_credit_score, min_months, min_rating, commission_rate, traffic_boost, benefits, create_time"

func toLevelTemplateProto(r *levelTemplateRow) *shop.ShopLevelTemplate {
	return &shop.ShopLevelTemplate{
		Level:          int32(r.Level),
		Name:           r.Name,
		MinGmv:         r.MinGmv,
		MinCreditScore: int32(r.MinCreditScore),
		MinMonths:      int32(r.MinMonths),
		MinRating:      r.MinRating,
		CommissionRate: r.CommissionRate,
		TrafficBoost:   r.TrafficBoost,
		Benefits:       r.Benefits,
	}
}

type levelApplicationRow struct {
	Id           uint64 `db:"id"`
	ShopId       int64  `db:"shop_id"`
	CurrentLevel int64  `db:"current_level"`
	TargetLevel  int64  `db:"target_level"`
	Snapshot     string `db:"snapshot"`
	Status       int64  `db:"status"`
	AdminId      int64  `db:"admin_id"`
	AdminRemark  string `db:"admin_remark"`
	CreateTime   int64  `db:"create_time"`
	UpdateTime   int64  `db:"update_time"`
}

const levelApplicationCols = "id, shop_id, current_level, target_level, snapshot, status, admin_id, admin_remark, create_time, update_time"

func toLevelApplicationProto(r *levelApplicationRow) *shop.ShopLevelApplication {
	return &shop.ShopLevelApplication{
		Id:           int64(r.Id),
		ShopId:       r.ShopId,
		CurrentLevel: int32(r.CurrentLevel),
		TargetLevel:  int32(r.TargetLevel),
		Snapshot:     r.Snapshot,
		Status:       int32(r.Status),
		AdminId:      r.AdminId,
		AdminRemark:  r.AdminRemark,
		CreateTime:   r.CreateTime,
	}
}
