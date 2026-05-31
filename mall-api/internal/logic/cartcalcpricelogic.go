// 购物车实时算价 - 调 PromotionRpc.CalcPrice, 返回 paid_amount + breakdown。
package logic

import (
	"context"
	"errors"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-promotion-rpc/promotionclient"
)

func CartCalcPrice(ctx context.Context, svcCtx *svc.ServiceContext, req *types.CartCalcPriceReq) (*types.CartCalcPriceResp, error) {
	userId := middleware.UidFromCtx(ctx)
	if userId <= 0 {
		return nil, errors.New("login required")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("items required")
	}

	items := make([]*promotionclient.CartItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &promotionclient.CartItem{
			SkuId: it.SkuId, ProductId: it.ProductId,
			ShopId: it.ShopId, CategoryId: it.CategoryId,
			OriginalPrice: it.OriginalPrice, Quantity: it.Quantity,
		})
	}

	resp, err := svcCtx.PromotionRpc.CalcPrice(ctx, &promotionclient.CalcPriceReq{
		UserId: userId, Items: items, CouponIds: req.CouponIds, ShippingFee: req.ShippingFee,
	})
	if err != nil {
		return nil, err
	}

	conflicts := make([]*types.PriceConflictDTO, 0, len(resp.Conflicts))
	for _, c := range resp.Conflicts {
		conflicts = append(conflicts, &types.PriceConflictDTO{CouponId: c.CouponId, Reason: c.Reason})
	}
	return &types.CartCalcPriceResp{
		TotalAmount:       resp.TotalAmount,
		PromotionDiscount: resp.PromotionDiscount,
		CouponDiscount:    resp.CouponDiscount,
		ShippingFee:       resp.ShippingFee,
		PaidAmount:        resp.PaidAmount,
		DiscountDetail:    resp.DiscountDetail,
		Conflicts:         conflicts,
	}, nil
}
