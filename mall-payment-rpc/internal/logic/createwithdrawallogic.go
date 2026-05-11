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

type CreateWithdrawalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateWithdrawalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawalLogic {
	return &CreateWithdrawalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateWithdrawal moves the requested amount from balance to frozen and
// inserts a pending withdrawal_request. Fails if balance is insufficient.
func (l *CreateWithdrawalLogic) CreateWithdrawal(in *payment.CreateWithdrawalReq) (*payment.CreateWithdrawalResp, error) {
	if in.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	now := time.Now().Unix()
	var newId int64
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		var balance int64
		if err := tx.QueryRowCtx(l.ctx, &balance,
			"SELECT balance FROM merchant_wallet WHERE shop_id = ? FOR UPDATE", in.ShopId); err != nil {
			return err
		}
		if balance < in.Amount {
			return errors.New("insufficient balance")
		}
		if _, err := tx.ExecCtx(l.ctx,
			"UPDATE merchant_wallet SET balance = balance - ?, frozen = frozen + ?, update_time = ? WHERE shop_id = ?",
			in.Amount, in.Amount, now, in.ShopId); err != nil {
			return err
		}
		res, err := tx.ExecCtx(l.ctx,
			"INSERT INTO withdrawal_request (shop_id, amount, bank_info, status, create_time, update_time) VALUES (?, ?, ?, 0, ?, ?)",
			in.ShopId, in.Amount, in.BankInfo, now, now)
		if err != nil {
			return err
		}
		newId, _ = res.LastInsertId()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &payment.CreateWithdrawalResp{Id: newId}, nil
}
