package logic

import (
	"context"
	"errors"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShopCouponStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShopCouponStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShopCouponStatusLogic {
	return &UpdateShopCouponStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateShopCouponStatusLogic) UpdateShopCouponStatus(in *activity.UpdateShopCouponStatusReq) (*activity.Empty, error) {
	if in.Id <= 0 || in.ShopId <= 0 {
		return nil, errors.New("id and shop_id required")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `shop_coupon` SET status=?, update_time=? WHERE id=? AND shop_id=?",
		in.Status, now, in.Id, in.ShopId,
	); err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
