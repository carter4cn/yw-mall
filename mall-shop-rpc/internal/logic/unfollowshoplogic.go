package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowShopLogic {
	return &UnfollowShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfollowShopLogic) UnfollowShop(in *shop.UnfollowShopReq) (*shop.OkResp, error) {
	// todo: add your logic here and delete this line

	return &shop.OkResp{}, nil
}
