package logic

import (
	"context"

	"mall-shop-rpc/internal/model"
	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRecommendedShopsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRecommendedShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRecommendedShopsLogic {
	return &ListRecommendedShopsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRecommendedShopsLogic) ListRecommendedShops(in *shop.ListRecommendedShopsReq) (*shop.ListShopsResp, error) {
	limit := in.Limit
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	var rows []*model.Shop
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, "SELECT * FROM shop WHERE status=1 ORDER BY rating DESC, follow_count DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: int64(len(out))}, nil
}
