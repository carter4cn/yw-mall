package logic

import (
	"context"
	"fmt"
	"strings"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLedgerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLedgerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLedgerLogic {
	return &ListLedgerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListLedger paginates account_ledger rows scoped by shop, optional category
// and optional time range. Order is create_time DESC, id DESC for stable
// pagination across same-second writes.
func (l *ListLedgerLogic) ListLedger(in *payment.ListLedgerReq) (*payment.ListLedgerResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	var (
		where strings.Builder
		args  []any
	)
	where.WriteString("1=1")
	if in.ShopId > 0 {
		where.WriteString(" AND shop_id = ?")
		args = append(args, in.ShopId)
	}
	if in.Category != "" {
		where.WriteString(" AND category = ?")
		args = append(args, in.Category)
	}
	if in.StartTime > 0 {
		where.WriteString(" AND create_time >= ?")
		args = append(args, in.StartTime)
	}
	if in.EndTime > 0 {
		where.WriteString(" AND create_time <= ?")
		args = append(args, in.EndTime)
	}

	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total,
		fmt.Sprintf("SELECT COUNT(*) FROM account_ledger WHERE %s", where.String()),
		args...,
	); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	var rows []ledgerEntryRow
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, pageSize, offset)
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows,
		fmt.Sprintf("SELECT %s FROM account_ledger WHERE %s ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?", ledgerColumns, where.String()),
		listArgs...,
	); err != nil {
		return nil, err
	}

	out := make([]*payment.LedgerEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, &payment.LedgerEntry{
			Id:             r.Id,
			ShopId:         r.ShopId,
			Direction:      int32(r.Direction),
			Category:       r.Category,
			Amount:         r.Amount,
			RunningBalance: r.RunningBalance,
			OrderId:        r.OrderId,
			RefundId:       r.RefundId,
			RefNo:          r.RefNo,
			Description:    r.Description,
			CreateTime:     r.CreateTime,
		})
	}
	return &payment.ListLedgerResp{Entries: out, Total: total}, nil
}
