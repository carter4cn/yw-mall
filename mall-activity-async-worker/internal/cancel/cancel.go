// Package cancel scans pending orders that have outlived the cashier TTL
// and marks them cancelled (status=4). Mirrors the shape of the settlement
// loop. Wallet rows are not touched because nothing was ever frozen for a
// pending order — the funds only flow on ConfirmMockPay / real callback.
package cancel

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

type Canceller struct {
	order      *sql.DB
	timeoutSec int
	tick       time.Duration
}

func New(orderDSN string, pendingTimeoutSec, tickSec int) (*Canceller, error) {
	od, err := sql.Open("mysql", orderDSN)
	if err != nil {
		return nil, fmt.Errorf("open order: %w", err)
	}
	return &Canceller{
		order:      od,
		timeoutSec: pendingTimeoutSec,
		tick:       time.Duration(tickSec) * time.Second,
	}, nil
}

// Run blocks until ctx is cancelled, scanning every tick.
func (c *Canceller) Run(ctx context.Context) {
	logx.Infof("cancel loop started: timeout=%ds tick=%s", c.timeoutSec, c.tick)
	t := time.NewTicker(c.tick)
	defer t.Stop()
	c.scanOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.scanOnce(ctx)
		}
	}
}

func (c *Canceller) scanOnce(ctx context.Context) {
	// Compare seconds since epoch on both sides to dodge any client/server
	// timezone mismatch when binding a Go time.Time → MySQL TIMESTAMP.
	res, err := c.order.ExecContext(ctx,
		"UPDATE `order` SET status = 4, cancel_time = UNIX_TIMESTAMP(), cancel_reason = 'auto:expired' "+
			"WHERE status = 0 AND cancel_time = 0 "+
			"AND UNIX_TIMESTAMP(create_time) + ? < UNIX_TIMESTAMP() "+
			"ORDER BY id LIMIT 200",
		c.timeoutSec,
	)
	if err != nil {
		logx.Errorf("cancel scan: %v", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		logx.Infof("cancel: expired %d pending orders", n)
	}
}
