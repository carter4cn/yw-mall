package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMerchantWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMerchantWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMerchantWalletLogic {
	return &GetMerchantWalletLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type walletRow struct {
	Id             int64 `db:"id"`
	ShopId         int64 `db:"shop_id"`
	Balance        int64 `db:"balance"`
	Frozen         int64 `db:"frozen"`
	TotalIncome    int64 `db:"total_income"`
	TotalWithdrawn int64 `db:"total_withdrawn"`
	CreateTime     int64 `db:"create_time"`
	UpdateTime     int64 `db:"update_time"`
}

const walletCols = "id, shop_id, balance, frozen, total_income, total_withdrawn, create_time, update_time"

// GetMerchantWallet upserts a zero-balance wallet for first-time shops.
func (l *GetMerchantWalletLogic) GetMerchantWallet(in *payment.GetMerchantWalletReq) (*payment.MerchantWallet, error) {
	var row walletRow
	err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &row,
		"SELECT "+walletCols+" FROM merchant_wallet WHERE shop_id = ?", in.ShopId)
	if errors.Is(err, sql.ErrNoRows) {
		now := time.Now().Unix()
		if _, ierr := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"INSERT INTO merchant_wallet (shop_id, balance, frozen, total_income, total_withdrawn, create_time, update_time) VALUES (?, 0, 0, 0, 0, ?, ?)",
			in.ShopId, now, now); ierr != nil {
			return nil, ierr
		}
		return &payment.MerchantWallet{ShopId: in.ShopId, UpdateTime: now}, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment.MerchantWallet{
		ShopId:         row.ShopId,
		Balance:        row.Balance,
		Frozen:         row.Frozen,
		TotalIncome:    row.TotalIncome,
		TotalWithdrawn: row.TotalWithdrawn,
		UpdateTime:     row.UpdateTime,
	}, nil
}
