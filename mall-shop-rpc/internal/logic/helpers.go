package logic

import (
	"mall-shop-rpc/internal/model"
	"mall-shop-rpc/shop"
)

func toShopProto(s *model.Shop) *shop.Shop {
	return &shop.Shop{
		Id:           int64(s.Id),
		Name:         s.Name,
		Logo:         s.Logo,
		Banner:       s.Banner,
		Description:  s.Description,
		Rating:       s.Rating,
		ProductCount: int32(s.ProductCount),
		FollowCount:  int32(s.FollowCount),
		Status:       int32(s.Status),
		CreateTime:   s.CreateTime,
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
