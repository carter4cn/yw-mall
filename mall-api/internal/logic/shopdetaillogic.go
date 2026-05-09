package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShopDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShopDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShopDetailLogic {
	return &ShopDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShopDetailLogic) ShopDetail(req *types.ShopDetailReq) (*types.ShopDetailResp, error) {
	res, err := l.svcCtx.ShopRpc.GetShop(l.ctx, &shopservice.GetShopReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	s := res.Shop
	return &types.ShopDetailResp{
		Shop: types.ShopItem{
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
		},
	}, nil
}
