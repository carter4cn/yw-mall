// Code scaffolded manually for S1.3 mock-confirm endpoint.

package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"

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
	return &types.OkResp{Ok: true}, nil
}
