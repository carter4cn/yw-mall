// Code scaffolded manually for S1.3 mock-confirm endpoint.

package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"
	"mall-promotion-rpc/promotionclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmMockPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmMockPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmMockPayLogic {
	return &ConfirmMockPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmMockPayLogic) ConfirmMockPay(req *types.ConfirmMockPayReq) (*types.OkResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	if _, err := l.svcCtx.PaymentRpc.ConfirmMockPay(l.ctx, &payment.ConfirmMockPayReq{
		OrderId: req.OrderId,
		UserId:  userId,
	}); err != nil {
		return nil, err
	}
	// Phase 1 S1.5.4: 支付成功 → 核销订单上锁定的券 (status 1→2)
	// best-effort, 失败 log 不阻塞支付成功响应（券核销失败不应让用户重复付款）
	if _, err := l.svcCtx.PromotionRpc.ConsumeCoupon(l.ctx, &promotionclient.ConsumeCouponReq{
		OrderId: req.OrderId,
	}); err != nil {
		logx.WithContext(l.ctx).Errorf("ConsumeCoupon orderId=%d failed: %v", req.OrderId, err)
	}
	return &types.OkResp{Ok: true}, nil
}
