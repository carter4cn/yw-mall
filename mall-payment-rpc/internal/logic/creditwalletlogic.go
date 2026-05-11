package logic

import (
	"context"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreditWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreditWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreditWalletLogic {
	return &CreditWalletLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreditWallet adds (positive amount) or deducts (negative amount) shop
// balance and writes a matching bill_record. Wallet row is upserted on first
// credit. total_income only accumulates positive credits.
func (l *CreditWalletLogic) CreditWallet(in *payment.CreditWalletReq) (*payment.OkResp, error) {
	now := time.Now().Unix()
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		incomeDelta := int64(0)
		if in.Amount > 0 {
			incomeDelta = in.Amount
		}
		if _, err := tx.ExecCtx(l.ctx,
			`INSERT INTO merchant_wallet (shop_id, balance, frozen, total_income, total_withdrawn, create_time, update_time)
			 VALUES (?, ?, 0, ?, 0, ?, ?)
			 ON DUPLICATE KEY UPDATE balance = balance + VALUES(balance), total_income = total_income + VALUES(total_income), update_time = VALUES(update_time)`,
			in.ShopId, in.Amount, incomeDelta, now, now); err != nil {
			return err
		}
		_, err := tx.ExecCtx(l.ctx,
			"INSERT INTO bill_record (shop_id, type, amount, order_id, remark, create_time) VALUES (?, ?, ?, ?, ?, ?)",
			in.ShopId, in.Type, in.Amount, in.OrderId, in.Remark, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &payment.OkResp{Ok: true}, nil
}
