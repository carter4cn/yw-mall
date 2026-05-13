package logic

import (
	"context"
	"errors"
	"time"

	"mall-payment-rpc/internal/channel"
	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ExecuteRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteRefundLogic {
	return &ExecuteRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteRefund performs the actual fund-callback step of a refund flow:
//
//  1. Calls the configured PayChannel (mock today) for confirmation.
//  2. Debits merchant_wallet.frozen (and total_income) by the refund amount,
//     guarded by a frozen >= amount check so the wallet never goes negative.
//  3. Writes a bill_record (type=refund, amount=-N) for ledger consistency.
//  4. Marks the matching payment row as status=2 (refunded) for traceability.
//
// All three writes share a single transaction so partial failure rolls back the
// wallet deduction; channel call happens outside the tx because it is the only
// idempotent external side-effect.
func (l *ExecuteRefundLogic) ExecuteRefund(in *payment.ExecuteRefundReq) (*payment.ExecuteRefundResp, error) {
	if in.Amount <= 0 || in.ShopId <= 0 {
		return nil, errors.New("invalid refund request")
	}

	// 1) Channel — default to mock if no DefaultChannel configured.
	channelName := l.svcCtx.Config.DefaultChannel
	if channelName == "" {
		channelName = "mock"
	}
	ch, ok := l.svcCtx.Channels[channelName]
	if !ok {
		var err error
		ch, err = channel.New(channelName)
		if err != nil {
			return nil, err
		}
	}
	chanResp, err := ch.Refund(l.ctx, &channel.RefundReq{
		OrderID: in.OrderId,
		OrderNo: in.OrderNo,
		Amount:  in.Amount,
		Reason:  in.Reason,
	})
	if err != nil {
		return nil, err
	}
	if chanResp.Status != "success" {
		return nil, errors.New("channel refund not successful: " + chanResp.Status)
	}

	refundNo := in.RefundNo
	if refundNo == "" {
		refundNo = chanResp.RefundNo
	}

	// 2) Wallet debit + bill record + payment status flip — all-or-nothing.
	now := time.Now().Unix()
	err = l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
		// Guarded deduction: only succeeds when frozen >= amount.
		res, err := tx.ExecCtx(ctx,
			"UPDATE merchant_wallet SET frozen = frozen - ?, total_income = total_income - ?, update_time = ? WHERE shop_id = ? AND frozen >= ?",
			in.Amount, in.Amount, now, in.ShopId, in.Amount,
		)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return errors.New("insufficient frozen balance for refund")
		}

		// Negative bill_record indicates deduction (P1 H convention).
		if _, err := tx.ExecCtx(ctx,
			"INSERT INTO bill_record (shop_id, type, amount, order_id, remark, create_time) VALUES (?, 'refund', ?, ?, ?, ?)",
			in.ShopId, -in.Amount, in.OrderId, refundNo, now,
		); err != nil {
			return err
		}

		// Best-effort payment status flip; skip if no row exists (early refund?).
		if _, err := tx.ExecCtx(ctx,
			"UPDATE payment SET status = 2 WHERE order_no = ? AND status = 1",
			in.OrderNo,
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &payment.ExecuteRefundResp{
		Success:  true,
		RefundNo: refundNo,
		Channel:  channelName,
	}, nil
}
