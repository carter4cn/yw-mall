package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShopListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShopListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShopListLogic {
	return &ShopListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShopListLogic) ShopList(req *types.ShopListReq) (*types.ShopListResp, error) {
	res, err := l.svcCtx.ShopRpc.ListShops(l.ctx, &shopservice.ListShopsReq{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
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
	return &types.ShopListResp{Shops: shops, Total: res.Total}, nil
}
