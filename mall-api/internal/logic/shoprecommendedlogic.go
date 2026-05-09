package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShopRecommendedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShopRecommendedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShopRecommendedLogic {
	return &ShopRecommendedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShopRecommendedLogic) ShopRecommended(req *types.ShopRecommendedReq) (*types.ShopRecommendedResp, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	res, err := l.svcCtx.ShopRpc.ListRecommendedShops(l.ctx, &shopservice.ListRecommendedShopsReq{Limit: limit})
	if err != nil {
		return nil, err
	}

	shops := make([]types.ShopItem, 0, len(res.Shops))
	for _, s := range res.Shops {
		shops = append(shops, types.ShopItem{
			Id:           s.Id,
			Name:         s.Name,
			Logo:         s.Logo,
			Banner:       s.Banner,
			Description:  s.Description,
			Rating:       s.Rating,
			ProductCount: s.ProductCount,
			FollowCount:  s.FollowCount,
			Status:       s.Status,
			CreateTime:   s.CreateTime,
		})
	}
	return &types.ShopRecommendedResp{Shops: shops}, nil
}
