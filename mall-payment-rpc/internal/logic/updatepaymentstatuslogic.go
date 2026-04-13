package logic

import (
	"context"
	"database/sql"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePaymentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePaymentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePaymentStatusLogic {
	return &UpdatePaymentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePaymentStatusLogic) UpdatePaymentStatus(in *payment.UpdatePaymentStatusReq) (*payment.UpdatePaymentStatusResp, error) {
	if in.Status == 1 {
		// success: set pay_time to now
		_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE payment SET status = ?, pay_time = ? WHERE id = ?",
			in.Status, sql.NullTime{Time: time.Now(), Valid: true}, in.Id)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE payment SET status = ? WHERE id = ?",
			in.Status, in.Id)
		if err != nil {
			return nil, err
		}
	}

	return &payment.UpdatePaymentStatusResp{}, nil
}
