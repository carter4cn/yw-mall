package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowedShopsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFollowedShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowedShopsLogic {
	return &ListFollowedShopsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFollowedShopsLogic) ListFollowedShops(req *types.ListFollowedShopsReq) (*types.ShopListResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.ShopRpc.ListFollowedShops(l.ctx, &shopservice.ListFollowedShopsReq{
		UserId:   userId,
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
