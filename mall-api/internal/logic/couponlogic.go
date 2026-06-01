// C 端券中心 + 我的券包 logic 层包装。
package logic

import (
	"context"
	"errors"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-promotion-rpc/promotionclient"
	"mall-user-rpc/userclient"
)

// ListAvailableCoupons C 端券中心 - 可领券列表。
// Phase 1 MVP: 返回所有 status=1 上架的券模板(按 shop_id 过滤 / 0=平台)。
// 后续版本扩展按"用户已领"过滤、新人券标记等。
func ListAvailableCoupons(ctx context.Context, svcCtx *svc.ServiceContext, req *types.CouponListReq) (*types.CouponListResp, error) {
	resp, err := svcCtx.PromotionRpc.ListCouponTemplates(ctx, &promotionclient.ListCouponTemplatesReq{
		ShopId: req.ShopId, Status: 1, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*types.CouponTemplateView, 0, len(resp.Templates))
	for _, t := range resp.Templates {
		out = append(out, &types.CouponTemplateView{
			TemplateId:    t.Id, ShopId: t.ShopId, Name: t.Name, Type: t.Type,
			Value: t.Value, MinAmount: t.MinAmount, MaxDiscount: t.MaxDiscount,
			TotalCount: t.TotalCount, ReceivedCount: t.ReceivedCount,
			PerUserLimit: t.PerUserLimit,
			ValidType: t.ValidType, ValidDays: t.ValidDays,
			ValidStart: t.ValidStart, ValidEnd: t.ValidEnd,
			ReceiveStart: t.ReceiveStart, ReceiveEnd: t.ReceiveEnd,
		})
	}
	return &types.CouponListResp{Total: resp.Total, Templates: out}, nil
}

// ReceiveCoupon 用户领券 - 自动带上 user.create_time 供新人券校验
func ReceiveCoupon(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ReceiveCouponReq) (*types.ReceiveCouponResp, error) {
	userId := middleware.UidFromCtx(ctx)
	if userId <= 0 {
		return nil, errors.New("login required")
	}
	// 查 user.create_time, 让 promotion-rpc 判断新人券资格
	var registerTime int64
	if u, err := svcCtx.UserRpc.GetUser(ctx, &userclient.GetUserReq{Id: userId}); err == nil && u != nil {
		registerTime = u.CreateTime
	}
	resp, err := svcCtx.PromotionRpc.ReceiveCoupon(ctx, &promotionclient.ReceiveCouponReq{
		UserId: userId, TemplateId: req.TemplateId, UserRegisterTime: registerTime,
	})
	if err != nil {
		return nil, err
	}
	return &types.ReceiveCouponResp{CouponId: resp.CouponId}, nil
}

// ListMyCoupons 我的券包 (按 status 过滤)。
func ListMyCoupons(ctx context.Context, svcCtx *svc.ServiceContext, req *types.MyCouponsReq) (*types.MyCouponsResp, error) {
	userId := middleware.UidFromCtx(ctx)
	if userId <= 0 {
		return nil, errors.New("login required")
	}
	resp, err := svcCtx.PromotionRpc.ListMyCoupons(ctx, &promotionclient.ListMyCouponsReq{
		UserId: userId, Status: req.Status, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*types.MyCouponView, 0, len(resp.Coupons))
	for _, c := range resp.Coupons {
		out = append(out, &types.MyCouponView{
			Id: c.Id, TemplateId: c.TemplateId, ShopId: c.ShopId, Status: c.Status,
			OrderId: c.OrderId, ReceiveTime: c.ReceiveTime, ExpireTime: c.ExpireTime,
			UseTime: c.UseTime,
			TemplateName: c.TemplateName, Type: c.Type, Value: c.Value, MinAmount: c.MinAmount,
		})
	}
	return &types.MyCouponsResp{Total: resp.Total, Coupons: out}, nil
}
