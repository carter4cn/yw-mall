package logic

import (
	"context"
	"errors"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserAppealRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserAppealRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserAppealRefundLogic {
	return &UserAppealRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserAppealRefund moves a rejected refund (status=2) into arbitration (status=3),
// recording the user's appeal reason and timestamp.
func (l *UserAppealRefundLogic) UserAppealRefund(in *order.UserAppealRefundReq) (*order.OkResp, error) {
	row, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if row.UserId != in.UserId {
		return nil, errors.New("refund does not belong to user")
	}
	if row.Status != 2 {
		return nil, errors.New("refund not in rejected state")
	}

	now := time.Now().Unix()
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET status = 3, appeal_reason = ?, appeal_time = ?, update_time = ? WHERE id = ? AND status = 2",
		in.Reason, now, now, in.RefundId,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("refund state changed concurrently")
	}
	return &order.OkResp{Ok: true}, nil
}
