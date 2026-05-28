package logic

import (
	"context"
	"errors"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShopDecorationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShopDecorationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShopDecorationLogic {
	return &UpdateShopDecorationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateShopDecoration UPSERT 模式: 没有就插, 有就改.
// 不做 owner 校验（admin-api 层用 perm decoration.write 把关）.
func (l *UpdateShopDecorationLogic) UpdateShopDecoration(in *shop.UpdateShopDecorationReq) (*shop.OkResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `
        INSERT INTO shop_decoration (shop_id, banners, announcement, featured_pids)
        VALUES (?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
          banners=VALUES(banners),
          announcement=VALUES(announcement),
          featured_pids=VALUES(featured_pids)`,
		in.ShopId, in.Banners, in.Announcement, in.FeaturedPids); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
