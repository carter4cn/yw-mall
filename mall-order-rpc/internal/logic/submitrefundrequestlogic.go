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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

	// 4) Build JSON payloads outside tx (无 DB IO，省锁内时间).
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

	// 整段事务化 + order 行锁 —— 防并发 submit 击穿 cumulative / pending 两个
	// 检查。同一订单的并发 SubmitRefund 会因 SELECT ... FOR UPDATE 串行化，
	// 第二个进入事务后看到第一个已写入的 refund_request 行就会立刻命中拒绝。
	now := time.Now().Unix()
	var (
		row orderRowForRefund
		id  int64
	)
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 1) Load order WITH row lock + validate ownership/state.
		if err := tx.QueryRowCtx(ctx, &row,
			"SELECT id, order_no, user_id, total_amount, status, shop_id FROM `order` WHERE id = ? FOR UPDATE",
			in.OrderId,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("order not found")
			}
			return err
		}
		if row.UserId != in.UserId {
			return errors.New("order does not belong to user")
		}
		// 已支付/已发货/已完成 才能申请退款；待支付(0)/已取消(4) 不允许。
		if row.Status != 1 && row.Status != 2 && row.Status != 3 {
			return errors.New("order not eligible for refund (must be paid/shipped/completed)")
		}

		// 2a) Pending dedup —— 同一订单只允许一个 active refund (status 0/1/3
		//     = 待审/审核中/申诉中)，避免商家面对两条"请处理"无所适从 + 防双扣。
		var activeCount int64
		if err := tx.QueryRowCtx(ctx, &activeCount,
			"SELECT COUNT(*) FROM refund_request WHERE order_id = ? AND status IN (0, 1, 3)",
			row.Id,
		); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if activeCount > 0 {
			return errors.New("已有进行中的退款申请，请先等待处理完成")
		}

		// 2b) Cumulative refund amount check (exclude rejected/final_rejected).
		var existingSum int64
		if err := tx.QueryRowCtx(ctx, &existingSum,
			"SELECT IFNULL(SUM(amount), 0) FROM refund_request WHERE order_id = ? AND status NOT IN (2, 5)",
			row.Id,
		); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if existingSum+in.Amount > row.TotalAmount {
			return errors.New("refund amount exceeds remaining order amount")
		}

		// 3) Rate-limit: max 3 submissions per order per 24h.
		var recentCount int64
		if err := tx.QueryRowCtx(ctx, &recentCount,
			"SELECT COUNT(*) FROM refund_request WHERE order_id = ? AND create_time > ?",
			row.Id, now-86400,
		); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if recentCount >= 3 {
			return errors.New("too many refund requests in 24h for this order")
		}

		// 5) INSERT refund_request.
		res, err := tx.ExecCtx(ctx,
			"INSERT INTO refund_request (order_id, order_no, user_id, shop_id, amount, reason, evidence, items, status, refund_type, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?)",
			row.Id, row.OrderNo, row.UserId, row.ShopId, in.Amount, in.Reason, evidenceJSON, itemsJSON, refundTypeOrDefault(in.RefundType), now, now,
		)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	if err != nil {
		return nil, err
	}

	// 6) S2.8 risk signal — log only, never block. Outside tx 不让指标失败拖垮主流程。
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
