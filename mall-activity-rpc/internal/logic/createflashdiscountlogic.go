package logic

import (
	"context"
	"errors"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFlashDiscountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFlashDiscountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFlashDiscountLogic {
	return &CreateFlashDiscountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateFlashDiscountLogic) CreateFlashDiscount(in *activity.CreateFlashDiscountReq) (*activity.CreateFlashDiscountResp, error) {
	if in.ShopId <= 0 || in.SkuId <= 0 {
		return nil, errors.New("shop_id and sku_id required")
	}
	if in.DiscountPrice <= 0 || in.OriginalPrice <= 0 {
		return nil, errors.New("price must be positive")
	}
	if in.DiscountPrice >= in.OriginalPrice {
		return nil, errors.New("discount_price must be less than original_price")
	}
	if in.StartTime >= in.EndTime {
		return nil, errors.New("start_time must be less than end_time")
	}
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO `sku_flash_discount`(shop_id, product_id, sku_id, original_price, discount_price, start_time, end_time, status, create_time, update_time) VALUES(?,?,?,?,?,?,?,1,?,?)",
		in.ShopId, in.ProductId, in.SkuId, in.OriginalPrice, in.DiscountPrice, in.StartTime, in.EndTime, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &activity.CreateFlashDiscountResp{Id: id}, nil
}
