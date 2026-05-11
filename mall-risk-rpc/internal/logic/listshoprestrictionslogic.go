package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopRestrictionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopRestrictionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopRestrictionsLogic {
	return &ListShopRestrictionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type shopRestrictionRow struct {
	Id          int64  `db:"id"`
	ShopId      int64  `db:"shop_id"`
	Restriction string `db:"restriction"`
	Reason      string `db:"reason"`
	OperatorId  int64  `db:"operator_id"`
	ExpireTime  int64  `db:"expire_time"`
	CreateTime  int64  `db:"create_time"`
}

func (l *ListShopRestrictionsLogic) ListShopRestrictions(in *risk.ListShopRestrictionsReq) (*risk.ListShopRestrictionsResp, error) {
	rows := []*shopRestrictionRow{}
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, shop_id, restriction, reason, operator_id, expire_time, create_time FROM shop_restriction WHERE shop_id = ? ORDER BY id DESC",
		in.ShopId); err != nil {
		return nil, err
	}
	out := make([]*risk.ShopRestriction, 0, len(rows))
	for _, r := range rows {
		out = append(out, &risk.ShopRestriction{
			Id:          r.Id,
			ShopId:      r.ShopId,
			Restriction: r.Restriction,
			Reason:      r.Reason,
			OperatorId:  r.OperatorId,
			ExpireTime:  r.ExpireTime,
			CreateTime:  r.CreateTime,
		})
	}
	return &risk.ListShopRestrictionsResp{Restrictions: out}, nil
}
