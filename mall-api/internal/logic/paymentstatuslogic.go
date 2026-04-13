// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentStatusLogic {
	return &PaymentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentStatusLogic) PaymentStatus(req *types.PaymentStatusReq) (resp *types.PaymentStatusResp, err error) {
	res, err := l.svcCtx.PaymentRpc.GetPayment(l.ctx, &payment.GetPaymentReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.PaymentStatusResp{
		Id:        res.Id,
		PaymentNo: res.PaymentNo,
		OrderNo:   res.OrderNo,
		Amount:    res.Amount,
		Status:    res.Status,
		PayType:   res.PayType,
		PayTime:   res.PayTime,
	}, nil
}
