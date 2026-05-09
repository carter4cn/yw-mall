package logic

import (
	"context"
	"time"

	"mall-common/errorx"
	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	if in.UserId == 0 || in.ShopId == 0 {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		res, e := sess.ExecCtx(ctx,
			"INSERT IGNORE INTO shop_follow(user_id, shop_id, create_time) VALUES (?, ?, ?)",
			in.UserId, in.ShopId, time.Now().Unix())
		if e != nil {
			return e
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return errorx.NewCodeError(errorx.ShopFollowAlreadyExists)
		}
		_, e = sess.ExecCtx(ctx, "UPDATE shop SET follow_count = follow_count + 1 WHERE id = ?", in.ShopId)
		return e
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
