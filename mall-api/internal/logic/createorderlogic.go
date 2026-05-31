// Phase 1 优惠活动接入版: 先 CalcPrice → 写订单 → LockCoupon
//
// 链路:
//   1. 拉默认地址
//   2. 反查每个 product 的 shop_id (engine 需要)
//   3. PromotionRpc.CalcPrice → 拿 paid_amount + discount_detail
//   4. order-rpc.CreateOrder (含 4 个优惠列)
//   5. PromotionRpc.LockCoupon (券 0→1)
//      失败回滚: 如果 LockCoupon 失败, 取消刚下的订单
package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-cart-rpc/cartclient"
	"mall-common/errorx"
	"mall-order-rpc/order"
	productpb "mall-product-rpc/product"
	"mall-promotion-rpc/promotionclient"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	userId := middleware.UidFromCtx(l.ctx)

	// 1) 默认地址
	addr, err := l.svcCtx.UserRpc.GetDefaultAddress(l.ctx, &userclient.GetDefaultAddressReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	if addr == nil || addr.Id == 0 {
		return nil, errorx.NewCodeErrorMsg(errorx.OrderAddressRequired, "请先添加默认收货地址")
	}

	// 2) 反查每个 product 的 shop_id (engine 需要 sku/shop)
	calcItems := make([]*promotionclient.CartItem, 0, len(req.Items))
	orderItems := make([]*order.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		p, perr := l.svcCtx.ProductRpc.GetProduct(l.ctx, &productpb.GetProductReq{Id: item.ProductId})
		if perr != nil {
			return nil, perr
		}
		calcItems = append(calcItems, &promotionclient.CartItem{
			SkuId:         item.ProductId, // Phase 1 sku_id = product_id (无 sku 子表场景)
			ProductId:     item.ProductId,
			ShopId:        p.ShopId,
			CategoryId:    p.CategoryId,
			OriginalPrice: item.Price,
			Quantity:      item.Quantity,
		})
		orderItems = append(orderItems, &order.OrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	// 3) CalcPrice
	calc, err := l.svcCtx.PromotionRpc.CalcPrice(l.ctx, &promotionclient.CalcPriceReq{
		UserId:      userId,
		Items:       calcItems,
		CouponIds:   req.CouponIds,
		ShippingFee: req.ShippingFee,
	})
	if err != nil {
		return nil, err
	}
	// 用户传了券但都不可用 → 拒绝下单
	if len(req.CouponIds) > 0 && calc.CouponDiscount == 0 && len(calc.Conflicts) > 0 {
		return nil, errorx.NewCodeErrorMsg(errorx.ParamError, "您选择的券不可用: "+calc.Conflicts[0].Reason)
	}

	// 4) 写订单 (含 4 个优惠字段)
	res, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &order.CreateOrderReq{
		UserId:            userId,
		AddressId:         addr.Id,
		Items:             orderItems,
		PromotionDiscount: calc.PromotionDiscount,
		CouponDiscount:    calc.CouponDiscount,
		PaidAmount:        calc.PaidAmount,
		DiscountDetail:    calc.DiscountDetail,
	})
	if err != nil {
		return nil, err
	}

	// 5) LockCoupon — 失败回滚刚下的订单
	if len(req.CouponIds) > 0 {
		if _, lerr := l.svcCtx.PromotionRpc.LockCoupon(l.ctx, &promotionclient.LockCouponReq{
			UserId: userId, OrderId: res.Id, CouponIds: req.CouponIds,
		}); lerr != nil {
			// 撤销订单 (设为已取消 status=4)
			_, _ = l.svcCtx.OrderRpc.UpdateOrderStatus(l.ctx, &order.UpdateOrderStatusReq{
				Id: res.Id, Status: 4,
			})
			return nil, lerr
		}
	}

	// 6) 下单成功 → 清理购物车里这些商品 (best-effort, 失败不阻塞返回)
	for _, item := range req.Items {
		if _, cerr := l.svcCtx.CartRpc.RemoveItem(l.ctx, &cartclient.RemoveItemReq{
			UserId: userId, ProductId: item.ProductId,
		}); cerr != nil {
			logx.WithContext(l.ctx).Errorf("RemoveCart userId=%d productId=%d failed: %v", userId, item.ProductId, cerr)
		}
	}

	return &types.CreateOrderResp{
		Id:          res.Id,
		OrderNo:     res.OrderNo,
		TotalAmount: res.TotalAmount,
	}, nil
}
