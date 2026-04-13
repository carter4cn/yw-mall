package logic

import (
	"context"

	"mall-payment-rpc/internal/model"
	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaymentLogic) GetPayment(in *payment.GetPaymentReq) (*payment.GetPaymentResp, error) {
	p, err := l.svcCtx.PaymentModel.FindOne(l.ctx, uint64(in.Id))
	if err == model.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var payTime int64
	if p.PayTime.Valid {
		payTime = p.PayTime.Time.Unix()
	}

	return &payment.GetPaymentResp{
		Id:        int64(p.Id),
		PaymentNo: p.PaymentNo,
		OrderNo:   p.OrderNo,
		UserId:    int64(p.UserId),
		Amount:    p.Amount,
		Status:    int32(p.Status),
		PayType:   int32(p.PayType),
		PayTime:   payTime,
	}, nil
}
