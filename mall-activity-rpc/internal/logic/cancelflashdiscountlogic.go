package logic

import (
	"context"
	"errors"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelFlashDiscountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelFlashDiscountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelFlashDiscountLogic {
	return &CancelFlashDiscountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelFlashDiscountLogic) CancelFlashDiscount(in *activity.CancelFlashDiscountReq) (*activity.Empty, error) {
	if in.Id <= 0 || in.ShopId <= 0 {
		return nil, errors.New("id and shop_id required")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `sku_flash_discount` SET status=2, update_time=? WHERE id=? AND shop_id=? AND status=1",
		now, in.Id, in.ShopId,
	); err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
