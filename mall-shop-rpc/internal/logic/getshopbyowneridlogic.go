package logic

import (
	"context"
	"errors"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetShopByOwnerIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopByOwnerIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopByOwnerIdLogic {
	return &GetShopByOwnerIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type shopDetailRow struct {
	Id              uint64  `db:"id"`
	Name            string  `db:"name"`
	Logo            string  `db:"logo"`
	Banner          string  `db:"banner"`
	Description     string  `db:"description"`
	Rating          float64 `db:"rating"`
	ProductCount    int64   `db:"product_count"`
	FollowCount     int64   `db:"follow_count"`
	Status          int64   `db:"status"`
	CreateTime      int64   `db:"create_time"`
	OwnerUserId     int64   `db:"owner_user_id"`
	CreditScore     int64   `db:"credit_score"`
	Level           int64   `db:"level"`
	ContactPhone    string  `db:"contact_phone"`
	BusinessLicense string  `db:"business_license"`
}

const shopDetailCols = "id, name, logo, banner, description, rating, product_count, follow_count, status, create_time, owner_user_id, credit_score, level, contact_phone, business_license"

func toShopDetailProto(r *shopDetailRow) *shop.ShopDetailResp {
	return &shop.ShopDetailResp{
		Id:              int64(r.Id),
		Name:            r.Name,
		Logo:            r.Logo,
		Banner:          r.Banner,
		Description:     r.Description,
		Rating:          r.Rating,
		ProductCount:    int32(r.ProductCount),
		FollowCount:     int32(r.FollowCount),
		Status:          int32(r.Status),
		CreateTime:      r.CreateTime,
		OwnerUserId:     r.OwnerUserId,
		CreditScore:     int32(r.CreditScore),
		Level:           int32(r.Level),
		ContactPhone:    r.ContactPhone,
		BusinessLicense: r.BusinessLicense,
	}
}

func (l *GetShopByOwnerIdLogic) GetShopByOwnerId(in *shop.GetShopByOwnerIdReq) (*shop.ShopDetailResp, error) {
	var r shopDetailRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &r,
		"SELECT "+shopDetailCols+" FROM shop WHERE owner_user_id=? LIMIT 1", in.OwnerUserId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("shop not found for owner")
		}
		return nil, err
	}
	return toShopDetailProto(&r), nil
}
