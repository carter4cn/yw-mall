package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopCouponsLogic {
	return &ListShopCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type shopCouponRow struct {
	Id              int64  `db:"id"`
	ShopId          int64  `db:"shop_id"`
	Code            string `db:"code"`
	Name            string `db:"name"`
	Type            int32  `db:"type"`
	DiscountValue   int64  `db:"discount_value"`
	MinOrderAmount  int64  `db:"min_order_amount"`
	TotalQuantity   int32  `db:"total_quantity"`
	ClaimedQuantity int32  `db:"claimed_quantity"`
	PerUserLimit    int32  `db:"per_user_limit"`
	ValidFrom       int64  `db:"valid_from"`
	ValidTo         int64  `db:"valid_to"`
	Status          int32  `db:"status"`
	CreateTime      int64  `db:"create_time"`
}

func (l *ListShopCouponsLogic) ListShopCoupons(in *activity.ListShopCouponsReq) (*activity.ListShopCouponsResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size
	clauses := []string{"shop_id=?"}
	args := []any{in.ShopId}
	if in.Status != 0 {
		clauses = append(clauses, "status=?")
		args = append(args, in.Status)
	}
	where := strings.Join(clauses, " AND ")
	rows := []shopCouponRow{}
	q := fmt.Sprintf("SELECT id, shop_id, code, name, type, discount_value, min_order_amount, total_quantity, claimed_quantity, per_user_limit, valid_from, valid_to, status, create_time FROM `shop_coupon` WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d", where, size, offset)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM `shop_coupon` WHERE %s", where), args...); err != nil {
		return nil, err
	}
	out := make([]*activity.ShopCoupon, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &activity.ShopCoupon{
			Id:              r.Id,
			ShopId:          r.ShopId,
			Code:            r.Code,
			Name:            r.Name,
			Type:            r.Type,
			DiscountValue:   r.DiscountValue,
			MinOrderAmount:  r.MinOrderAmount,
			TotalQuantity:   r.TotalQuantity,
			ClaimedQuantity: r.ClaimedQuantity,
			PerUserLimit:    r.PerUserLimit,
			ValidFrom:       r.ValidFrom,
			ValidTo:         r.ValidTo,
			Status:          r.Status,
			CreateTime:      r.CreateTime,
		})
	}
	return &activity.ListShopCouponsResp{Coupons: out, Total: total}, nil
}
