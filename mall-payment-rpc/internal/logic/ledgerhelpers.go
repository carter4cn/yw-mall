package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ledgerEntryRow is the package-level projection of an account_ledger row used
// by the ListLedger / GetShopLedgerSummary readers.
type ledgerEntryRow struct {
	Id             int64  `db:"id"`
	ShopId         int64  `db:"shop_id"`
	Direction      int64  `db:"direction"`
	Category       string `db:"category"`
	Amount         int64  `db:"amount"`
	RunningBalance int64  `db:"running_balance"`
	OrderId        int64  `db:"order_id"`
	RefundId       int64  `db:"refund_id"`
	WithdrawalId   int64  `db:"withdrawal_id"`
	RefNo          string `db:"ref_no"`
	Description    string `db:"description"`
	CreateTime     int64  `db:"create_time"`
}

const ledgerColumns = "id, shop_id, direction, category, amount, running_balance, order_id, refund_id, withdrawal_id, ref_no, description, create_time"

// writeLedgerEntry inserts a single account_ledger row inside the caller's
// transaction. The running_balance is computed from the latest wallet snapshot
// (balance+frozen) so reconciliation can compare ledger_net vs wallet_total.
// direction: 1=credit, 2=debit. amount must always be positive.
func writeLedgerEntry(ctx context.Context, tx sqlx.Session, shopId int64, direction int8, category string, amount int64, orderId, refundId, withdrawalId int64, refNo, description string) error {
	if shopId <= 0 || amount <= 0 {
		return nil
	}
	var walletTotal int64
	if err := tx.QueryRowCtx(ctx, &walletTotal,
		"SELECT IFNULL(balance + frozen, 0) FROM merchant_wallet WHERE shop_id = ? LIMIT 1",
		shopId,
	); err != nil {
		// If wallet row doesn't exist yet, fall through with 0.
		walletTotal = 0
	}
	now := time.Now().Unix()
	_, err := tx.ExecCtx(ctx,
		"INSERT INTO account_ledger (shop_id, direction, category, amount, running_balance, order_id, refund_id, withdrawal_id, ref_no, description, create_time) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		shopId, direction, category, amount, walletTotal, orderId, refundId, withdrawalId, refNo, description, now,
	)
	return err
}
