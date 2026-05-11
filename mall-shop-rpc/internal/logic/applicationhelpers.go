package logic

import (
	"mall-shop-rpc/shop"
)

type applicationRow struct {
	Id              uint64 `db:"id"`
	UserId          int64  `db:"user_id"`
	ShopName        string `db:"shop_name"`
	Logo            string `db:"logo"`
	Description     string `db:"description"`
	ContactPhone    string `db:"contact_phone"`
	BusinessLicense string `db:"business_license"`
	LegalPerson     string `db:"legal_person"`
	IdCardFront     string `db:"id_card_front"`
	IdCardBack      string `db:"id_card_back"`
	Category        string `db:"category"`
	Status          int64  `db:"status"`
	ReviewRemark    string `db:"review_remark"`
	ReviewerId      int64  `db:"reviewer_id"`
	ShopId          int64  `db:"shop_id"`
	CreateTime      int64  `db:"create_time"`
	UpdateTime      int64  `db:"update_time"`
}

func toApplicationProto(r *applicationRow) *shop.ShopApplication {
	return &shop.ShopApplication{
		Id:              int64(r.Id),
		UserId:          r.UserId,
		ShopName:        r.ShopName,
		Logo:            r.Logo,
		Description:     r.Description,
		ContactPhone:    r.ContactPhone,
		BusinessLicense: r.BusinessLicense,
		LegalPerson:     r.LegalPerson,
		IdCardFront:     r.IdCardFront,
		IdCardBack:      r.IdCardBack,
		Category:        r.Category,
		Status:          int32(r.Status),
		ReviewRemark:    r.ReviewRemark,
		ReviewerId:      r.ReviewerId,
		ShopId:          r.ShopId,
		CreateTime:      r.CreateTime,
		UpdateTime:      r.UpdateTime,
	}
}

const applicationCols = "id, user_id, shop_name, logo, description, contact_phone, business_license, legal_person, id_card_front, id_card_back, category, status, review_remark, reviewer_id, shop_id, create_time, update_time"
