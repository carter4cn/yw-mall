package logic

import (
	"context"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunReconciliationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunReconciliationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunReconciliationLogic {
	return &RunReconciliationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RunReconciliation compares per-shop ledger_net (sum credit - sum debit) with
// merchant_wallet.balance+frozen. shop_id=0 → all shops. A diff outside ±1 cent
// flips passed=false. The report is meant for ops dashboards / nightly cron.
func (l *RunReconciliationLogic) RunReconciliation(in *payment.RunReconciliationReq) (*payment.ReconciliationReport, error) {
	// 1) Determine target shop ids.
	type shopRow struct {
		ShopId       int64 `db:"shop_id"`
		Balance      int64 `db:"balance"`
		Frozen       int64 `db:"frozen"`
		TotalIncome  int64 `db:"total_income"`
	}
	var wallets []shopRow
	if in.ShopId > 0 {
		if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &wallets,
			"SELECT shop_id, balance, frozen, total_income FROM merchant_wallet WHERE shop_id = ?",
			in.ShopId,
		); err != nil {
			return nil, err
		}
	} else {
		if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &wallets,
			"SELECT shop_id, balance, frozen, total_income FROM merchant_wallet ORDER BY shop_id",
		); err != nil {
			return nil, err
		}
	}

	report := &payment.ReconciliationReport{}
	for _, w := range wallets {
		type sumRow struct {
			Direction int64 `db:"direction"`
			Total     int64 `db:"total"`
		}
		var sums []sumRow
		if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &sums,
			"SELECT direction, SUM(amount) AS total FROM account_ledger WHERE shop_id = ? GROUP BY direction",
			w.ShopId,
		); err != nil {
			return nil, err
		}
		var credit, debit int64
		for _, s := range sums {
			if s.Direction == 1 {
				credit = s.Total
			} else if s.Direction == 2 {
				debit = s.Total
			}
		}
		net := credit - debit
		walletTotal := w.Balance + w.Frozen
		diff := net - walletTotal
		passed := diff >= -1 && diff <= 1

		report.Results = append(report.Results, &payment.ShopReconcileResult{
			ShopId:        w.ShopId,
			LedgerCredit:  credit,
			LedgerDebit:   debit,
			LedgerNet:     net,
			WalletBalance: w.Balance,
			WalletFrozen:  w.Frozen,
			WalletTotal:   walletTotal,
			Diff:          diff,
			Passed:        passed,
		})
		report.TotalChecked++
		if passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}
	return report, nil
}
