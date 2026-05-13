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

type MerchantHandleRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMerchantHandleRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantHandleRefundLogic {
	return &MerchantHandleRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MerchantHandleRefund processes a merchant's accept/reject decision on a pending
// refund request. action=1 approves and triggers ExecuteRefund (payment-rpc) →
// flips status to 4 (refunded). action=2 rejects with remark → status 2.
func (l *MerchantHandleRefundLogic) MerchantHandleRefund(in *order.MerchantHandleRefundReq) (*order.OkResp, error) {
	if in.Action != 1 && in.Action != 2 {
		return nil, errors.New("invalid action")
	}

	row, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if row.ShopId != in.ShopId {
		return nil, errors.New("refund does not belong to shop")
	}
	if row.Status != 0 {
		return nil, errors.New("refund not in pending state")
	}

	now := time.Now().Unix()
	if in.Action == 2 {
		// reject
		res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 2, merchant_user_id = ?, merchant_remark = ?, merchant_handle_time = ?, update_time = ? WHERE id = ? AND status = 0",
			in.MerchantUserId, in.Remark, now, now, in.RefundId,
		)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("refund state changed concurrently")
		}
		return &order.OkResp{Ok: true}, nil
	}

	// action=1: approve + execute refund
	refundNo := generateRefundNo()
	// 1) optimistic flip pending -> approved
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET status = 1, merchant_user_id = ?, merchant_remark = ?, merchant_handle_time = ?, refund_no = ?, update_time = ? WHERE id = ? AND status = 0",
		in.MerchantUserId, in.Remark, now, refundNo, now, in.RefundId,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("refund state changed concurrently")
	}

	// 2) call payment-rpc ExecuteRefund (wallet debit + bill_record + payment status update)
	if l.svcCtx.PaymentRpc == nil {
		// rollback approval to avoid stuck state
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 0, refund_no = '' WHERE id = ? AND status = 1",
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
		// rollback approval so merchant can retry / admin can intervene
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 0, refund_no = '' WHERE id = ? AND status = 1",
			in.RefundId,
		)
		return nil, err
	}

	// 3) mark refunded
	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET status = 4, refund_complete_time = ?, update_time = ? WHERE id = ? AND status = 1",
		now, now, in.RefundId,
	); err != nil {
		return nil, err
	}
	return &order.OkResp{Ok: true}, nil
}
