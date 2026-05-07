package logic

import (
	"context"

	"mall-shop-rpc/internal/model"
	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowedShopsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowedShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowedShopsLogic {
	return &ListFollowedShopsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowedShopsLogic) ListFollowedShops(in *shop.ListFollowedShopsReq) (*shop.ListShopsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size
	var rows []*model.Shop
	q := `SELECT s.* FROM shop s
	      INNER JOIN shop_follow f ON s.id = f.shop_id
	      WHERE f.user_id = ? AND s.status = 1
	      ORDER BY f.create_time DESC LIMIT ?, ?`
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, in.UserId, offset, size); err != nil {
		return nil, err
	}
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: int64(len(out))}, nil
}
