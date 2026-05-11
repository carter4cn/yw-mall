package logic

import (
	"context"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShopStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShopStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShopStatusLogic {
	return &UpdateShopStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateShopStatusLogic) UpdateShopStatus(in *shop.UpdateShopStatusReq) (*shop.OkResp, error) {
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE shop SET status=?, update_time=? WHERE id=?",
		in.Status, now, in.ShopId); err != nil {
		return nil, err
	}
	if in.Reason != "" {
		l.Logger.Infof("shop %d status changed to %d reason=%s", in.ShopId, in.Status, in.Reason)
	}
	return &shop.OkResp{Ok: true}, nil
}
