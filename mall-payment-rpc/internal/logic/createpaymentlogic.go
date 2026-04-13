package logic

import (
	"context"
	"fmt"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePaymentLogic) CreatePayment(in *payment.CreatePaymentReq) (*payment.CreatePaymentResp, error) {
	paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	result, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO payment (payment_no, order_no, user_id, amount, status, pay_type) VALUES (?, ?, ?, ?, 0, ?)",
		paymentNo, in.OrderNo, in.UserId, in.Amount, in.PayType)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &payment.CreatePaymentResp{Id: id, PaymentNo: paymentNo}, nil
}
