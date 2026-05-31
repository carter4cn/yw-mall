package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

const couponTemplateCols = "id, activity_id, shop_id, name, type, value, min_amount, max_discount, category_id, total_count, received_count, used_count, per_user_limit, valid_type, valid_days, valid_start, valid_end, receive_start, receive_end, status, create_time, update_time"

type couponTemplateRow struct {
	Id            int64     `db:"id"`
	ActivityId    int64     `db:"activity_id"`
	ShopId        int64     `db:"shop_id"`
	Name          string    `db:"name"`
	Type          string    `db:"type"`
	Value         int64     `db:"value"`
	MinAmount     int64     `db:"min_amount"`
	MaxDiscount   int64     `db:"max_discount"`
	CategoryId    int64     `db:"category_id"`
	TotalCount    int32     `db:"total_count"`
	ReceivedCount int32     `db:"received_count"`
	UsedCount     int32     `db:"used_count"`
	PerUserLimit  int32     `db:"per_user_limit"`
	ValidType     int32     `db:"valid_type"`
	ValidDays     int32     `db:"valid_days"`
	ValidStart    int64     `db:"valid_start"`
	ValidEnd      int64     `db:"valid_end"`
	ReceiveStart  int64     `db:"receive_start"`
	ReceiveEnd    int64     `db:"receive_end"`
	Status        int32     `db:"status"`
	CreateTime    time.Time `db:"create_time"`
	UpdateTime    time.Time `db:"update_time"`
}

func couponTemplateRowToPb(r *couponTemplateRow) *promotion.CouponTemplate {
	return &promotion.CouponTemplate{
		Id: r.Id, ActivityId: r.ActivityId, ShopId: r.ShopId, Name: r.Name, Type: r.Type,
		Value: r.Value, MinAmount: r.MinAmount, MaxDiscount: r.MaxDiscount, CategoryId: r.CategoryId,
		TotalCount: r.TotalCount, ReceivedCount: r.ReceivedCount, UsedCount: r.UsedCount,
		PerUserLimit: r.PerUserLimit, ValidType: r.ValidType, ValidDays: r.ValidDays,
		ValidStart: r.ValidStart, ValidEnd: r.ValidEnd,
		ReceiveStart: r.ReceiveStart, ReceiveEnd: r.ReceiveEnd, Status: r.Status,
		CreateTime: r.CreateTime.Unix(), UpdateTime: r.UpdateTime.Unix(),
	}
}

// ListCouponTemplates 按 shop_id (0=平台) + status 过滤分页查询。
func ListCouponTemplates(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ListCouponTemplatesReq) (*promotion.ListCouponTemplatesResp, error) {
	if in.ShopId < 0 {
		return nil, errors.New("shop_id required")
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	where := []string{"shop_id = ?"}
	args := []any{in.ShopId}
	if in.Status >= 0 {
		where = append(where, "status = ?")
		args = append(args, in.Status)
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	if err := svcCtx.DB.QueryRowCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(*) FROM coupon_template WHERE %s", whereClause), args...); err != nil {
		return nil, err
	}

	var rows []couponTemplateRow
	listArgs := append(args, pageSize, (page-1)*pageSize)
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		fmt.Sprintf("SELECT %s FROM coupon_template WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", couponTemplateCols, whereClause),
		listArgs...); err != nil {
		return nil, err
	}

	out := make([]*promotion.CouponTemplate, 0, len(rows))
	for i := range rows {
		out = append(out, couponTemplateRowToPb(&rows[i]))
	}
	return &promotion.ListCouponTemplatesResp{Templates: out, Total: total}, nil
}

// ChangeCouponTemplateStatus 简单 status 切换 (0下架/1上架)。
func ChangeCouponTemplateStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ChangeCouponTemplateStatusReq) (*promotion.OkResp, error) {
	if in.Status != 0 && in.Status != 1 {
		return nil, errors.New("invalid status")
	}
	if _, err := svcCtx.DB.ExecCtx(ctx,
		"UPDATE coupon_template SET status = ? WHERE id = ?", in.Status, in.Id,
	); err != nil {
		return nil, err
	}
	return &promotion.OkResp{Ok: true}, nil
}
