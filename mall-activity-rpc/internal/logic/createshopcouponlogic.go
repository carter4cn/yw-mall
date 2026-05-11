package logic

import (
	"context"
	"errors"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShopCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShopCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShopCouponLogic {
	return &CreateShopCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateShopCouponLogic) CreateShopCoupon(in *activity.CreateShopCouponReq) (*activity.CreateShopCouponResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	if in.Type < 1 || in.Type > 3 {
		return nil, errors.New("type must be 1=满减,2=折扣,3=直减")
	}
	if in.ValidTo <= in.ValidFrom {
		return nil, errors.New("valid_to must be greater than valid_from")
	}
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO `shop_coupon`(shop_id, code, name, type, discount_value, min_order_amount, total_quantity, claimed_quantity, per_user_limit, valid_from, valid_to, status, create_time, update_time) VALUES(?,?,?,?,?,?,?,0,?,?,?,1,?,?)",
		in.ShopId, in.Code, in.Name, in.Type, in.DiscountValue, in.MinOrderAmount, in.TotalQuantity, in.PerUserLimit, in.ValidFrom, in.ValidTo, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &activity.CreateShopCouponResp{Id: id}, nil
}
