package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFollowingLogic {
	return &IsFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsFollowingLogic) IsFollowing(in *shop.IsFollowingReq) (*shop.IsFollowingResp, error) {
	var n int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &n, "SELECT COUNT(*) FROM shop_follow WHERE user_id=? AND shop_id=?", in.UserId, in.ShopId)
	if err != nil {
		return nil, err
	}
	return &shop.IsFollowingResp{IsFollowing: n > 0}, nil
}
