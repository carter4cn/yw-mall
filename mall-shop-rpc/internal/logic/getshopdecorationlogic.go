package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetShopDecorationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopDecorationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopDecorationLogic {
	return &GetShopDecorationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetShopDecoration 返回该店的装修配置, 未配置时返空字段 (ShopId 仍带回)
func (l *GetShopDecorationLogic) GetShopDecoration(in *shop.GetShopDecorationReq) (*shop.ShopDecoration, error) {
	var row struct {
		ShopId       uint64 `db:"shop_id"`
		Banners      string `db:"banners"`
		Announcement string `db:"announcement"`
		FeaturedPids string `db:"featured_pids"`
		UpdateTime   int64  `db:"update_time_unix"`
	}
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row, `
        SELECT shop_id, banners, announcement, featured_pids,
               UNIX_TIMESTAMP(update_time) AS update_time_unix
        FROM shop_decoration WHERE shop_id=? LIMIT 1`, in.ShopId)
	if err == sqlx.ErrNotFound {
		return &shop.ShopDecoration{ShopId: in.ShopId}, nil
	}
	if err != nil {
		return nil, err
	}
	return &shop.ShopDecoration{
		ShopId:       int64(row.ShopId),
		Banners:      row.Banners,
		Announcement: row.Announcement,
		FeaturedPids: row.FeaturedPids,
		UpdateTime:   row.UpdateTime,
	}, nil
}
