package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyShopLogic {
	return &ApplyShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApplyShopLogic) ApplyShop(in *shop.ApplyShopReq) (*shop.ApplyShopResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("user_id required")
	}
	if strings.TrimSpace(in.ShopName) == "" {
		return nil, errors.New("shop_name required")
	}
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO shop_application (user_id, shop_name, logo, description, contact_phone, business_license, legal_person, id_card_front, id_card_back, category, status, review_remark, reviewer_id, shop_id, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, '', 0, 0, ?, ?)`,
		in.UserId, in.ShopName, in.Logo, in.Description, in.ContactPhone, in.BusinessLicense, in.LegalPerson, in.IdCardFront, in.IdCardBack, in.Category, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &shop.ApplyShopResp{ApplicationId: id}, nil
}
