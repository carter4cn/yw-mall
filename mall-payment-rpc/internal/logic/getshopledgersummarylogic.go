package logic

import (
	"context"
	"errors"
	"strings"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShopLedgerSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopLedgerSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopLedgerSummaryLogic {
	return &GetShopLedgerSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetShopLedgerSummary returns a single-shop ledger summary aggregated by
// category over an optional time window. net_balance = sum(credit) - sum(debit).
func (l *GetShopLedgerSummaryLogic) GetShopLedgerSummary(in *payment.GetShopLedgerSummaryReq) (*payment.LedgerSummary, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}

	var (
		where strings.Builder
		args  []any
	)
	where.WriteString("shop_id = ?")
	args = append(args, in.ShopId)
	if in.StartTime > 0 {
		where.WriteString(" AND create_time >= ?")
		args = append(args, in.StartTime)
	}
	if in.EndTime > 0 {
		where.WriteString(" AND create_time <= ?")
		args = append(args, in.EndTime)
	}

	type row struct {
		Category  string `db:"category"`
		Direction int64  `db:"direction"`
		Total     int64  `db:"total"`
	}
	var rows []row
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows,
		"SELECT category, direction, SUM(amount) AS total FROM account_ledger WHERE "+where.String()+" GROUP BY category, direction",
		args...,
	); err != nil {
		return nil, err
	}

	summary := &payment.LedgerSummary{}
	var credit, debit int64
	for _, r := range rows {
		if r.Direction == 1 {
			credit += r.Total
		} else if r.Direction == 2 {
			debit += r.Total
		}
		switch r.Category {
		case "order_income":
			if r.Direction == 1 {
				summary.TotalIncome += r.Total
			}
		case "refund":
			if r.Direction == 2 {
				summary.TotalRefund += r.Total
			}
		case "commission":
			if r.Direction == 2 {
				summary.TotalCommission += r.Total
			}
		case "withdrawal":
			if r.Direction == 2 {
				summary.TotalWithdrawal += r.Total
			}
		}
	}
	summary.NetBalance = credit - debit
	return summary, nil
}
