package logic

import (
	"context"
	"strconv"
	"strings"

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

	// 装修是软依赖 —— 拿不到就返空，不阻塞店铺详情。商家未配置装修时
	// shop-rpc 返 ShopId=0 的空对象（GetShopDecoration 内部对 NotFound 兜底）。
	var banners []string
	var announcement string
	var featuredPids []int64
	deco, derr := l.svcCtx.ShopRpc.GetShopDecoration(l.ctx, &shopservice.GetShopDecorationReq{ShopId: req.Id})
	if derr != nil {
		logx.WithContext(l.ctx).Errorf("ShopDetail: GetShopDecoration shopId=%d failed: %v", req.Id, derr)
	} else if deco != nil && deco.ShopId > 0 {
		if deco.Banners != "" {
			banners = strings.Split(deco.Banners, ",")
		}
		announcement = deco.Announcement
		if deco.FeaturedPids != "" {
			for _, p := range strings.Split(deco.FeaturedPids, ",") {
				if pid, perr := strconv.ParseInt(strings.TrimSpace(p), 10, 64); perr == nil && pid > 0 {
					featuredPids = append(featuredPids, pid)
				}
			}
		}
	}

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
		Banners:      banners,
		Announcement: announcement,
		FeaturedPids: featuredPids,
	}, nil
}
