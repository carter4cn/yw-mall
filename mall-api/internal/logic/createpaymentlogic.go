// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (resp *types.CreatePaymentResp, err error) {
	userId := middleware.UidFromCtx(l.ctx)

	res, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment.CreatePaymentReq{
		OrderNo: req.OrderNo,
		UserId:  userId,
		Amount:  req.Amount,
		PayType: req.PayType,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreatePaymentResp{
		Id:        res.Id,
		PaymentNo: res.PaymentNo,
	}, nil
}
