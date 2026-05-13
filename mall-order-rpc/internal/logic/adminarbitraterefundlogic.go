package logic

import (
	"context"
	"errors"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"
	"mall-payment-rpc/paymentclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminArbitrateRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminArbitrateRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminArbitrateRefundLogic {
	return &AdminArbitrateRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AdminArbitrateRefund finalizes an arbitrating (status=3) refund. action=1
// force-refunds (debit wallet → status=4). action=2 final-rejects → status=5.
//
// TODO(S2.4): auto-escalate refunds where the merchant has not responded within
// 72h. The cron worker will live in mall-activity-async-worker; not in this
// sprint.
func (l *AdminArbitrateRefundLogic) AdminArbitrateRefund(in *order.AdminArbitrateRefundReq) (*order.OkResp, error) {
	if in.Action != 1 && in.Action != 2 {
		return nil, errors.New("invalid action")
	}
	row, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if row.Status != 3 {
		return nil, errors.New("refund not in arbitrating state")
	}

	now := time.Now().Unix()
	if in.Action == 2 {
		res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 5, admin_id = ?, admin_remark = ?, admin_handle_time = ?, update_time = ? WHERE id = ? AND status = 3",
			in.AdminId, in.Remark, now, now, in.RefundId,
		)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("refund state changed concurrently")
		}
		return &order.OkResp{Ok: true}, nil
	}

	// action=1 force_refund
	refundNo := row.RefundNo
	if refundNo == "" {
		refundNo = generateRefundNo()
	}
	// move arbitrating -> approved (1) with admin stamp + refund_no
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET status = 1, admin_id = ?, admin_remark = ?, admin_handle_time = ?, refund_no = ?, update_time = ? WHERE id = ? AND status = 3",
		in.AdminId, in.Remark, now, refundNo, now, in.RefundId,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("refund state changed concurrently")
	}

	if l.svcCtx.PaymentRpc == nil {
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 3 WHERE id = ? AND status = 1",
			in.RefundId,
		)
		return nil, errors.New("payment rpc not configured")
	}
	if _, err := l.svcCtx.PaymentRpc.ExecuteRefund(l.ctx, &paymentclient.ExecuteRefundReq{
		OrderId:  row.OrderId,
		OrderNo:  row.OrderNo,
		ShopId:   row.ShopId,
		Amount:   row.Amount,
		Reason:   row.Reason,
		RefundNo: refundNo,
	}); err != nil {
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 3 WHERE id = ? AND status = 1",
			in.RefundId,
		)
		return nil, err
	}

	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET status = 4, refund_complete_time = ?, update_time = ? WHERE id = ? AND status = 1",
		now, now, in.RefundId,
	); err != nil {
		return nil, err
	}
	return &order.OkResp{Ok: true}, nil
}
