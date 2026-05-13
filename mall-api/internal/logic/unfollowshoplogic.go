package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowShopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfollowShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowShopLogic {
	return &UnfollowShopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfollowShopLogic) UnfollowShop(req *types.FollowShopReq) (*types.FollowShopResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	_, err := l.svcCtx.ShopRpc.UnfollowShop(l.ctx, &shopservice.UnfollowShopReq{
		UserId: userId,
		ShopId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.FollowShopResp{Ok: true}, nil
}
