package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowShopLogic {
	return &FollowShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowShopLogic) FollowShop(in *shop.FollowShopReq) (*shop.OkResp, error) {
	// todo: add your logic here and delete this line

	return &shop.OkResp{}, nil
}
