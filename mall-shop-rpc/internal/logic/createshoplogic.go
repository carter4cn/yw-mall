package logic

import (
	"context"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShopLogic {
	return &CreateShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateShopLogic) CreateShop(in *shop.CreateShopReq) (*shop.CreateShopResp, error) {
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO shop (name, logo, banner, description, rating, product_count, follow_count, status, create_time, update_time) VALUES (?, ?, ?, ?, ?, 0, 0, 1, ?, ?)",
		in.Name, in.Logo, in.Banner, in.Description, in.Rating, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &shop.CreateShopResp{Id: id}, nil
}
