package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		res, e := sess.ExecCtx(ctx, "DELETE FROM shop_follow WHERE user_id=? AND shop_id=?", in.UserId, in.ShopId)
		if e != nil {
			return e
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil
		}
		_, e = sess.ExecCtx(ctx, "UPDATE shop SET follow_count = GREATEST(follow_count - 1, 0) WHERE id = ?", in.ShopId)
		return e
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
