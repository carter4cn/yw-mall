// Package settlement runs the periodic T+N order settlement loop.
//
// It scans completed orders past the configured cooling-off window and
// transfers the order amount into the merchant wallet (with a bill record),
// flipping order.settle_status from 0 to 1 atomically to guarantee idempotency.
package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

type Settler struct {
	order   *sql.DB
	payment *sql.DB
	delay   time.Duration
	tick    time.Duration
}

func New(orderDSN, paymentDSN string, delaySec, tickSec int) (*Settler, error) {
	od, err := sql.Open("mysql", orderDSN)
	if err != nil {
		return nil, fmt.Errorf("open order: %w", err)
	}
	pd, err := sql.Open("mysql", paymentDSN)
	if err != nil {
		return nil, fmt.Errorf("open payment: %w", err)
	}
	return &Settler{
		order:   od,
		payment: pd,
		delay:   time.Duration(delaySec) * time.Second,
		tick:    time.Duration(tickSec) * time.Second,
	}, nil
}

// Run blocks until ctx is cancelled, scanning every tick.
func (s *Settler) Run(ctx context.Context) {
	logx.Infof("settlement loop started: delay=%s tick=%s", s.delay, s.tick)
	t := time.NewTicker(s.tick)
	defer t.Stop()
	s.scanOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.scanOnce(ctx)
		}
	}
}

type pendingOrder struct {
	ID      int64
	ShopID  int64
	Amount  int64
	OrderNo string
}

func (s *Settler) scanOnce(ctx context.Context) {
	cutoff := time.Now().Add(-s.delay).Unix()
	rows, err := s.order.QueryContext(ctx, "SELECT id, shop_id, total_amount, order_no FROM `order` "+
		"WHERE status = 3 AND settle_status = 0 AND refund_status IN (0, 3) "+
		"AND complete_time > 0 AND complete_time <= ? LIMIT 200", cutoff)
	if err != nil {
		logx.Errorf("settle scan: %v", err)
		return
	}
	var pending []pendingOrder
	for rows.Next() {
		var p pendingOrder
		if err := rows.Scan(&p.ID, &p.ShopID, &p.Amount, &p.OrderNo); err != nil {
			logx.Errorf("settle row scan: %v", err)
			continue
		}
		pending = append(pending, p)
	}
	rows.Close()

	for _, p := range pending {
		if err := s.settleOne(ctx, p); err != nil {
			logx.Errorf("settle order %d: %v", p.ID, err)
		}
	}
	if len(pending) > 0 {
		logx.Infof("settle: processed %d orders", len(pending))
	}
}

func (s *Settler) settleOne(ctx context.Context, p pendingOrder) error {
	if p.ShopID == 0 || p.Amount <= 0 {
		_, _ = s.order.ExecContext(ctx,
			"UPDATE `order` SET settle_status = 2 WHERE id = ? AND settle_status = 0", p.ID)
		return nil
	}
	// CAS claim — guarantees only one worker settles each order.
	res, err := s.order.ExecContext(ctx,
		"UPDATE `order` SET settle_status = 1 WHERE id = ? AND settle_status = 0", p.ID)
	if err != nil {
		return fmt.Errorf("claim: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}

	tx, err := s.payment.BeginTx(ctx, nil)
	if err != nil {
		s.releaseClaim(ctx, p.ID)
		return fmt.Errorf("begin tx: %w", err)
	}
	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO merchant_wallet (shop_id, balance, total_income, create_time, update_time) "+
			"VALUES (?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE balance = balance + VALUES(balance), "+
			"total_income = total_income + VALUES(total_income), update_time = VALUES(update_time)",
		p.ShopID, p.Amount, p.Amount, now, now,
	); err != nil {
		tx.Rollback()
		s.releaseClaim(ctx, p.ID)
		return fmt.Errorf("wallet upsert: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO bill_record (shop_id, type, amount, order_id, remark, create_time) "+
			"VALUES (?, 'settle', ?, ?, ?, ?)",
		p.ShopID, p.Amount, p.ID, fmt.Sprintf("T+settle order_no=%s", p.OrderNo), now,
	); err != nil {
		tx.Rollback()
		s.releaseClaim(ctx, p.ID)
		return fmt.Errorf("bill insert: %w", err)
	}
	if err := tx.Commit(); err != nil {
		s.releaseClaim(ctx, p.ID)
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Settler) releaseClaim(ctx context.Context, id int64) {
	_, _ = s.order.ExecContext(ctx,
		"UPDATE `order` SET settle_status = 0 WHERE id = ? AND settle_status = 1", id)
}
