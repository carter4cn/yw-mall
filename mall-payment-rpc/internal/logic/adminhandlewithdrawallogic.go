package logic

import (
	"context"
	"errors"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminHandleWithdrawalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminHandleWithdrawalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminHandleWithdrawalLogic {
	return &AdminHandleWithdrawalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AdminHandleWithdrawal completes a pending withdrawal:
//   action=1 (approve)  → status=3 paid, frozen-=amount, total_withdrawn+=amount,
//                         bill_record(type=withdrawal, amount=-amount)
//   action=2 (reject)   → status=2 rejected, frozen-=amount, balance+=amount
func (l *AdminHandleWithdrawalLogic) AdminHandleWithdrawal(in *payment.AdminHandleWithdrawalReq) (*payment.OkResp, error) {
	if in.Action != 1 && in.Action != 2 {
		return nil, errors.New("invalid action")
	}
	now := time.Now().Unix()
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		var (
			shopId int64
			amount int64
			status int64
		)
		row := struct {
			ShopId int64 `db:"shop_id"`
			Amount int64 `db:"amount"`
			Status int64 `db:"status"`
		}{}
		if err := tx.QueryRowCtx(l.ctx, &row,
			"SELECT shop_id, amount, status FROM withdrawal_request WHERE id = ? FOR UPDATE", in.Id); err != nil {
			return err
		}
		shopId, amount, status = row.ShopId, row.Amount, row.Status
		if status != 0 {
			return errors.New("withdrawal already handled")
		}

		if in.Action == 1 {
			if _, err := tx.ExecCtx(l.ctx,
				"UPDATE withdrawal_request SET status=3, admin_id=?, admin_remark=?, update_time=? WHERE id=?",
				in.AdminId, in.Remark, now, in.Id); err != nil {
				return err
			}
			if _, err := tx.ExecCtx(l.ctx,
				"UPDATE merchant_wallet SET frozen = frozen - ?, total_withdrawn = total_withdrawn + ?, update_time = ? WHERE shop_id = ?",
				amount, amount, now, shopId); err != nil {
				return err
			}
			if _, err := tx.ExecCtx(l.ctx,
				"INSERT INTO bill_record (shop_id, type, amount, order_id, remark, create_time) VALUES (?, 'withdrawal', ?, 0, ?, ?)",
				shopId, -amount, in.Remark, now); err != nil {
				return err
			}
			return nil
		}

		// reject: refund frozen back to balance
		if _, err := tx.ExecCtx(l.ctx,
			"UPDATE withdrawal_request SET status=2, admin_id=?, admin_remark=?, update_time=? WHERE id=?",
			in.AdminId, in.Remark, now, in.Id); err != nil {
			return err
		}
		_, err := tx.ExecCtx(l.ctx,
			"UPDATE merchant_wallet SET frozen = frozen - ?, balance = balance + ?, update_time = ? WHERE shop_id = ?",
			amount, amount, now, shopId)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &payment.OkResp{Ok: true}, nil
}
