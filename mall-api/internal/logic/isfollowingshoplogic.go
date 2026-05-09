package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsFollowingShopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsFollowingShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFollowingShopLogic {
	return &IsFollowingShopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsFollowingShopLogic) IsFollowingShop(req *types.IsFollowingReq) (*types.IsFollowingResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.ShopRpc.IsFollowing(l.ctx, &shopservice.IsFollowingReq{
		UserId: userId,
		ShopId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.IsFollowingResp{Following: res.IsFollowing}, nil
}
