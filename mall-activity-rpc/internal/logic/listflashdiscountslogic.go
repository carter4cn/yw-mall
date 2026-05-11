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

type ListFlashDiscountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFlashDiscountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFlashDiscountsLogic {
	return &ListFlashDiscountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type flashDiscountRow struct {
	Id            int64 `db:"id"`
	ShopId        int64 `db:"shop_id"`
	ProductId     int64 `db:"product_id"`
	SkuId         int64 `db:"sku_id"`
	OriginalPrice int64 `db:"original_price"`
	DiscountPrice int64 `db:"discount_price"`
	StartTime     int64 `db:"start_time"`
	EndTime       int64 `db:"end_time"`
	Status        int32 `db:"status"`
	CreateTime    int64 `db:"create_time"`
}

func (l *ListFlashDiscountsLogic) ListFlashDiscounts(in *activity.ListFlashDiscountsReq) (*activity.ListFlashDiscountsResp, error) {
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
	rows := []flashDiscountRow{}
	q := fmt.Sprintf("SELECT id, shop_id, product_id, sku_id, original_price, discount_price, start_time, end_time, status, create_time FROM `sku_flash_discount` WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d", where, size, offset)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM `sku_flash_discount` WHERE %s", where), args...); err != nil {
		return nil, err
	}
	out := make([]*activity.FlashDiscount, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &activity.FlashDiscount{
			Id:            r.Id,
			ShopId:        r.ShopId,
			ProductId:     r.ProductId,
			SkuId:         r.SkuId,
			OriginalPrice: r.OriginalPrice,
			DiscountPrice: r.DiscountPrice,
			StartTime:     r.StartTime,
			EndTime:       r.EndTime,
			Status:        r.Status,
			CreateTime:    r.CreateTime,
		})
	}
	return &activity.ListFlashDiscountsResp{Discounts: out, Total: total}, nil
}
