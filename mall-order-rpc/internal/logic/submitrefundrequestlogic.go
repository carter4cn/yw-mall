package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitRefundRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitRefundRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitRefundRequestLogic {
	return &SubmitRefundRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SubmitRefundRequest creates a new refund_request row in status=0 (pending).
// Validates order ownership + payable status + cumulative amount within order.total_amount.
// Caps 24h-per-order requests to 3 to deter abuse, and emits an F-5 risk log when
// the shop's 24h refund rate exceeds 10% (S2.8) — non-blocking.
func (l *SubmitRefundRequestLogic) SubmitRefundRequest(in *order.SubmitRefundRequestReq) (*order.SubmitRefundRequestResp, error) {
	if in.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	// 1) Load order + validate ownership/state.
	var row orderRowForRefund
	err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &row,
		"SELECT id, order_no, user_id, total_amount, status, shop_id FROM `order` WHERE id = ? LIMIT 1",
		in.OrderId,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}
	if row.UserId != in.UserId {
		return nil, errors.New("order does not belong to user")
	}
	// 已支付/已发货/已完成 才能申请退款；待支付(0)/已取消(4) 不允许。
	if row.Status != 1 && row.Status != 2 && row.Status != 3 {
		return nil, errors.New("order not eligible for refund (must be paid/shipped/completed)")
	}

	// 2) Cumulative refund amount check (exclude rejected/final_rejected).
	var existingSum int64
	if err = l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &existingSum,
		"SELECT IFNULL(SUM(amount), 0) FROM refund_request WHERE order_id = ? AND status NOT IN (2, 5)",
		row.Id,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingSum+in.Amount > row.TotalAmount {
		return nil, errors.New("refund amount exceeds remaining order amount")
	}

	// 3) Rate-limit: max 3 submissions per order per 24h.
	now := time.Now().Unix()
	var recentCount int64
	if err = l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &recentCount,
		"SELECT COUNT(*) FROM refund_request WHERE order_id = ? AND create_time > ?",
		row.Id, now-86400,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if recentCount >= 3 {
		return nil, errors.New("too many refund requests in 24h for this order")
	}

	// 4) Build JSON payloads.
	evidenceJSON := "[]"
	if len(in.Evidence) > 0 {
		if b, jerr := json.Marshal(in.Evidence); jerr == nil {
			evidenceJSON = string(b)
		}
	}
	itemsJSON := "[]"
	if len(in.Items) > 0 {
		simplified := make([]map[string]any, 0, len(in.Items))
		for _, it := range in.Items {
			simplified = append(simplified, map[string]any{
				"skuId":    it.SkuId,
				"skuName":  it.SkuName,
				"quantity": it.Quantity,
				"amount":   it.Amount,
			})
		}
		if b, jerr := json.Marshal(simplified); jerr == nil {
			itemsJSON = string(b)
		}
	}

	// 5) INSERT refund_request.
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO refund_request (order_id, order_no, user_id, shop_id, amount, reason, evidence, items, status, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)",
		row.Id, row.OrderNo, row.UserId, row.ShopId, in.Amount, in.Reason, evidenceJSON, itemsJSON, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 6) S2.8 risk signal — log only, never block.
	l.evaluateShopRefundRate(row.ShopId)

	return &order.SubmitRefundRequestResp{RefundId: id}, nil
}

// evaluateShopRefundRate computes the 24h refund-rate for a shop and logs an F-5
// alert if the rate exceeds 10%. Strictly observational; failures are swallowed.
func (l *SubmitRefundRequestLogic) evaluateShopRefundRate(shopId int64) {
	if shopId <= 0 {
		return
	}
	since := time.Now().Unix() - 86400
	var paidOrders, refundCount int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &paidOrders,
		"SELECT COUNT(*) FROM `order` WHERE shop_id = ? AND pay_time > ?",
		shopId, since,
	); err != nil {
		return
	}
	if paidOrders == 0 {
		return
	}
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &refundCount,
		"SELECT COUNT(*) FROM refund_request WHERE shop_id = ? AND create_time > ?",
		shopId, since,
	); err != nil {
		return
	}
	rate := float64(refundCount) / float64(paidOrders)
	if rate > 0.10 {
		logx.Errorf("F-5 alert: shop %d refund_rate=%.4f exceeds threshold (refunds=%d paid_orders=%d)",
			shopId, rate, refundCount, paidOrders)
	}
}
