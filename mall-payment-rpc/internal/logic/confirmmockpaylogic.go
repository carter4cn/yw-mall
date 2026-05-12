package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConfirmMockPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmMockPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmMockPayLogic {
	return &ConfirmMockPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ConfirmMockPay simulates a successful payment for status=0 orders when the
// PaymentMockEnabled flag is on. It flips order.status to 1, stamps pay_time,
// writes a payment row (pay_type=0 mock), and freezes the amount on the shop
// merchant_wallet (balance stays untouched until T+N settlement).
func (l *ConfirmMockPayLogic) ConfirmMockPay(in *payment.ConfirmMockPayReq) (*payment.OkResp, error) {
	if !l.svcCtx.Config.PaymentMockEnabled {
		return nil, status.Error(codes.PermissionDenied, "mock pay disabled")
	}
	if l.svcCtx.OrderDB == nil {
		return nil, errors.New("order datasource not configured")
	}

	// 1) Read order from mall_order to validate ownership + state + amount + shop.
	var row struct {
		Id          int64  `db:"id"`
		OrderNo     string `db:"order_no"`
		UserId      int64  `db:"user_id"`
		TotalAmount int64  `db:"total_amount"`
		Status      int64  `db:"status"`
		ShopId      int64  `db:"shop_id"`
	}
	err := l.svcCtx.OrderDB.QueryRowCtx(l.ctx, &row,
		"SELECT id, order_no, user_id, total_amount, status, shop_id FROM `order` WHERE id = ? LIMIT 1",
		in.OrderId,
	)
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

	// 2) CAS-flip order to paid + stamp pay_time. If RowsAffected==0 someone
	//    else already paid / cancelled — bail out safely.
	now := time.Now().Unix()
	res, err := l.svcCtx.OrderDB.ExecCtx(l.ctx,
		"UPDATE `order` SET status = 1, pay_time = ? WHERE id = ? AND status = 0",
		now, row.Id,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("order state changed concurrently")
	}

	// 3) Insert payment row + bump merchant_wallet.frozen in one tx. Failures
	//    here log and rollback order.status to keep ledger consistent.
	paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
	err = l.svcCtx.SqlConn.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		if _, ierr := tx.ExecCtx(l.ctx,
			"INSERT INTO payment (payment_no, order_no, user_id, amount, status, pay_type, pay_time) VALUES (?, ?, ?, ?, 1, 0, NOW())",
			paymentNo, row.OrderNo, row.UserId, row.TotalAmount,
		); ierr != nil {
			return ierr
		}
		if row.ShopId > 0 && row.TotalAmount > 0 {
			if _, ierr := tx.ExecCtx(l.ctx,
				`INSERT INTO merchant_wallet (shop_id, balance, frozen, total_income, total_withdrawn, create_time, update_time)
				 VALUES (?, 0, ?, 0, 0, ?, ?)
				 ON DUPLICATE KEY UPDATE frozen = frozen + VALUES(frozen), update_time = VALUES(update_time)`,
				row.ShopId, row.TotalAmount, now, now,
			); ierr != nil {
				return ierr
			}
		}
		return nil
	})
	if err != nil {
		// Roll order back so user can retry.
		_, _ = l.svcCtx.OrderDB.ExecCtx(l.ctx,
			"UPDATE `order` SET status = 0, pay_time = 0 WHERE id = ? AND status = 1",
			row.Id,
		)
		return nil, err
	}
	return &payment.OkResp{Ok: true}, nil
}
