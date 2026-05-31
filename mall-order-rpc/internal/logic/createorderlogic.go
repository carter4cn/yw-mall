package logic

import (
	"context"
	"errors"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/internal/util"
	"mall-order-rpc/order"
	"mall-product-rpc/productclient"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	addr, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &userclient.GetAddressReq{
		UserId: in.UserId,
		Id:     in.AddressId,
	})
	if err != nil {
		return nil, err
	}

	orderNo := util.GenerateOrderNo()

	var totalAmount int64
	for _, item := range in.Items {
		totalAmount += item.Price * int64(item.Quantity)
	}

	// 反查 shop_id —— 用首件商品的归属店铺，写到 order.shop_id；否则商家
	// 后台按 shop_id 过滤会查不到该订单（财务流水、订单管理都是同一过滤逻辑）。
	// 当前单一店铺购物车模型；多店铺购物车需另行拆单，校验所有 item 同店。
	if len(in.Items) == 0 {
		return nil, errors.New("订单不能为空")
	}
	firstProd, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &productclient.GetProductReq{
		Id: in.Items[0].ProductId,
	})
	if err != nil {
		return nil, err
	}
	shopId := firstProd.ShopId
	// 防御：所有 item 必须同店
	for _, item := range in.Items[1:] {
		p, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &productclient.GetProductReq{Id: item.ProductId})
		if err != nil {
			return nil, err
		}
		if p.ShopId != shopId {
			return nil, errors.New("同一订单的商品必须属于同一店铺")
		}
	}

	// Phase 1 优惠字段：mall-api 提前调 CalcPrice 算好后传入；
	// 若全空（无活动 / 老调用方）则 paid_amount 兜底 = totalAmount，discount_detail = null。
	paidAmount := in.PaidAmount
	if paidAmount <= 0 {
		paidAmount = totalAmount
	}
	var discountDetail any = nil
	if in.DiscountDetail != "" {
		discountDetail = in.DiscountDetail
	}

	var orderId int64
	err = l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			"INSERT INTO `order` (`order_no`, `user_id`, `shop_id`, `total_amount`, `promotion_discount`, `coupon_discount`, `paid_amount`, `discount_detail`, `status`, `address_id`, `receiver_name`, `receiver_phone`, `receiver_province`, `receiver_city`, `receiver_district`, `receiver_detail`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			orderNo, in.UserId, shopId, totalAmount,
			in.PromotionDiscount, in.CouponDiscount, paidAmount, discountDetail,
			0, addr.Id, addr.ReceiverName, addr.Phone,
			addr.Province, addr.City, addr.District, addr.Detail,
		)
		if err != nil {
			return err
		}

		orderId, err = result.LastInsertId()
		if err != nil {
			return err
		}

		for _, item := range in.Items {
			_, err = session.ExecCtx(ctx,
				"INSERT INTO `order_item` (`order_id`, `product_id`, `product_name`, `price`, `quantity`) VALUES (?, ?, ?, ?, ?)",
				orderId, item.ProductId, item.ProductName, item.Price, item.Quantity,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order.CreateOrderResp{
		Id:          orderId,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
	}, nil
}
