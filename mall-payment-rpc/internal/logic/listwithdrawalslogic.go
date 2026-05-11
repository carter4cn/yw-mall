package logic

import (
	"context"
	"strings"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawalsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawalsLogic {
	return &ListWithdrawalsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type withdrawalRow struct {
	Id          int64  `db:"id"`
	ShopId      int64  `db:"shop_id"`
	Amount      int64  `db:"amount"`
	BankInfo    string `db:"bank_info"`
	Status      int64  `db:"status"`
	AdminId     int64  `db:"admin_id"`
	AdminRemark string `db:"admin_remark"`
	CreateTime  int64  `db:"create_time"`
	UpdateTime  int64  `db:"update_time"`
}

const withdrawalCols = "id, shop_id, amount, bank_info, status, admin_id, admin_remark, create_time, update_time"

// ListWithdrawals supports both merchant and admin views. shop_id=0 means
// "all shops" (admin only). status=-1 means any status.
func (l *ListWithdrawalsLogic) ListWithdrawals(in *payment.ListWithdrawalsReq) (*payment.ListWithdrawalsResp, error) {
	conds := []string{}
	args := []any{}
	if in.ShopId > 0 {
		conds = append(conds, "shop_id = ?")
		args = append(args, in.ShopId)
	}
	if in.Status >= 0 {
		conds = append(conds, "status = ?")
		args = append(args, in.Status)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM withdrawal_request "+where, args...); err != nil {
		return nil, err
	}

	page, pageSize := in.Page, in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, pageSize, offset)

	rows := []*withdrawalRow{}
	q := "SELECT " + withdrawalCols + " FROM withdrawal_request " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows, q, pagedArgs...); err != nil {
		return nil, err
	}

	out := make([]*payment.WithdrawalInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, &payment.WithdrawalInfo{
			Id:          r.Id,
			ShopId:      r.ShopId,
			Amount:      r.Amount,
			BankInfo:    r.BankInfo,
			Status:      int32(r.Status),
			AdminId:     r.AdminId,
			AdminRemark: r.AdminRemark,
			CreateTime:  r.CreateTime,
			UpdateTime:  r.UpdateTime,
		})
	}
	return &payment.ListWithdrawalsResp{Withdrawals: out, Total: total}, nil
}
