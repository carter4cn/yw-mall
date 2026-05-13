// Code scaffolded manually for S1.2 cashier endpoint.

package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCashierLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCashierLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCashierLogic {
	return &GetCashierLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCashierLogic) GetCashier(req *types.GetCashierReq) (*types.CashierInfoResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	res, err := l.svcCtx.PaymentRpc.GetCashier(l.ctx, &payment.GetCashierReq{
		OrderId: req.OrderId,
		UserId:  userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.CashierInfoResp{
		OrderId:     res.OrderId,
		OrderNo:     res.OrderNo,
		Amount:      res.Amount,
		ExpireAt:    res.ExpireAt,
		Channels:    res.Channels,
		MockEnabled: res.MockEnabled,
	}, nil
}
