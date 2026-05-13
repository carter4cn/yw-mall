package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowShopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowShopLogic {
	return &FollowShopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FollowShopLogic) FollowShop(req *types.FollowShopReq) (*types.FollowShopResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	_, err := l.svcCtx.ShopRpc.FollowShop(l.ctx, &shopservice.FollowShopReq{
		UserId: userId,
		ShopId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.FollowShopResp{Ok: true}, nil
}
