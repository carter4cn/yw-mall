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

// ListMyCoupons 用户券包查询。
// status=-1 全部 / 0 未用 / 1 已锁定 / 2 已使用 / 3 已过期
// 副作用：每次查询前先把 expire_time < now 的 status=0 标记成 3 (过期)。
func ListMyCoupons(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ListMyCouponsReq) (*promotion.ListMyCouponsResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("user_id required")
	}
	now := time.Now().Unix()
	// 先把过期券归档（best-effort, 失败不阻塞查询）
	_, _ = svcCtx.DB.ExecCtx(ctx,
		"UPDATE coupon SET status = 3 WHERE user_id = ? AND status = 0 AND expire_time < ?",
		in.UserId, now,
	)

	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	where := []string{"c.user_id = ?"}
	args := []any{in.UserId}
	if in.Status >= 0 {
		where = append(where, "c.status = ?")
		args = append(args, in.Status)
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	if err := svcCtx.DB.QueryRowCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(*) FROM coupon c WHERE %s", whereClause), args...); err != nil {
		return nil, err
	}

	type row struct {
		Id           int64  `db:"id"`
		TemplateId   int64  `db:"template_id"`
		UserId       int64  `db:"user_id"`
		ShopId       int64  `db:"shop_id"`
		Status       int32  `db:"status"`
		OrderId      int64  `db:"order_id"`
		ReceiveTime  int64  `db:"receive_time"`
		ExpireTime   int64  `db:"expire_time"`
		LockTime     int64  `db:"lock_time"`
		UseTime      int64  `db:"use_time"`
		TemplateName string `db:"template_name"`
		Type         string `db:"type"`
		Value        int64  `db:"value"`
		MinAmount    int64  `db:"min_amount"`
	}
	q := fmt.Sprintf(`
		SELECT c.id, c.template_id, c.user_id, c.shop_id, c.status, c.order_id,
		       c.receive_time, c.expire_time, c.lock_time, c.use_time,
		       t.name AS template_name, t.type, t.value, t.min_amount
		FROM coupon c JOIN coupon_template t ON t.id = c.template_id
		WHERE %s
		ORDER BY c.id DESC LIMIT ? OFFSET ?`, whereClause)
	listArgs := append(args, pageSize, (page-1)*pageSize)
	var rows []row
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows, q, listArgs...); err != nil {
		return nil, err
	}

	out := make([]*promotion.Coupon, 0, len(rows))
	for _, r := range rows {
		out = append(out, &promotion.Coupon{
			Id: r.Id, TemplateId: r.TemplateId, UserId: r.UserId, ShopId: r.ShopId,
			Status: r.Status, OrderId: r.OrderId,
			ReceiveTime: r.ReceiveTime, ExpireTime: r.ExpireTime,
			LockTime: r.LockTime, UseTime: r.UseTime,
			TemplateName: r.TemplateName, Type: r.Type, Value: r.Value, MinAmount: r.MinAmount,
		})
	}
	return &promotion.ListMyCouponsResp{Coupons: out, Total: total}, nil
}
