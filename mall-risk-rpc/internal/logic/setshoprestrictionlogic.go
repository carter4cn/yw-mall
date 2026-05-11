package logic

import (
	"context"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SetShopRestrictionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetShopRestrictionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetShopRestrictionLogic {
	return &SetShopRestrictionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetShopRestriction upserts a restriction by deleting any existing record
// for the same shop_id+restriction pair before inserting the new one.
func (l *SetShopRestrictionLogic) SetShopRestriction(in *risk.SetShopRestrictionReq) (*risk.Empty, error) {
	now := time.Now().Unix()
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		if _, err := tx.ExecCtx(l.ctx,
			"DELETE FROM shop_restriction WHERE shop_id=? AND restriction=?",
			in.ShopId, in.Restriction); err != nil {
			return err
		}
		_, err := tx.ExecCtx(l.ctx,
			"INSERT INTO shop_restriction (shop_id, restriction, reason, operator_id, expire_time, create_time) VALUES (?, ?, ?, ?, ?, ?)",
			in.ShopId, in.Restriction, in.Reason, in.OperatorId, in.ExpireTime, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &risk.Empty{}, nil
}
