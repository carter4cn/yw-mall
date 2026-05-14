package logic

import (
	"context"
	"database/sql"
	"errors"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const pendingOrderTTLSec = 15 * 60 // S1.4 / S1.2 default cashier expiry

type GetCashierLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCashierLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCashierLogic {
	return &GetCashierLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCashier returns cashier metadata for a pending order. Verifies user
// ownership and pending status; returns the channels currently allowed plus
// the mock-enabled flag (S1.8 feature flag).
func (l *GetCashierLogic) GetCashier(in *payment.GetCashierReq) (*payment.CashierInfo, error) {
	if l.svcCtx.OrderDB == nil {
		return nil, errors.New("order datasource not configured")
	}
	var row struct {
		Id          int64  `db:"id"`
		OrderNo     string `db:"order_no"`
		UserId      int64  `db:"user_id"`
		TotalAmount int64  `db:"total_amount"`
		Status      int64  `db:"status"`
		CreateTime  int64  `db:"create_time"`
	}
	// Wrap in TransactCtx to pin the read to ProxySQL's default hostgroup
	// (master, hg=10). Without this, ProxySQL routes plain SELECTs to slaves
	// (hg=20) which lag 1-2s behind master, so a cashier load fired
	// immediately after order creation will fail with "order not found".
	err := l.svcCtx.OrderDB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		return sess.QueryRowCtx(ctx, &row,
			"SELECT id, order_no, user_id, total_amount, status, UNIX_TIMESTAMP(create_time) AS create_time FROM `order` WHERE id = ? LIMIT 1",
			in.OrderId,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}
	if row.UserId != in.UserId {
		return nil, errors.New("order does not belong to user")
	}
	if row.Status != 0 {
		return nil, errors.New("order not in pending state")
	}

	channels := []string{"mock"}
	return &payment.CashierInfo{
		OrderId:     row.Id,
		OrderNo:     row.OrderNo,
		Amount:      row.TotalAmount,
		ExpireAt:    row.CreateTime + pendingOrderTTLSec,
		Channels:    channels,
		MockEnabled: l.svcCtx.Config.PaymentMockEnabled,
	}, nil
}
