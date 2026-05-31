package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

// ListActivities 按 shop_id (必填) + type/status (可选) 分页查询。
// 列表项不返回 targets/actions 减小 payload；详情页用 GetActivity。
func ListActivities(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ListActivitiesReq) (*promotion.ListActivitiesResp, error) {
	if in.ShopId < 0 {
		return nil, errors.New("shop_id required (0 for platform)")
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
	if in.Type != "" {
		where = append(where, "type = ?")
		args = append(args, in.Type)
	}
	if in.Status >= 0 {
		where = append(where, "status = ?")
		args = append(args, in.Status)
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	if err := svcCtx.DB.QueryRowCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(*) FROM activity WHERE %s", whereClause), args...); err != nil {
		return nil, err
	}

	var rows []activityRow
	listArgs := append(args, pageSize, (page-1)*pageSize)
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		fmt.Sprintf("SELECT %s FROM activity WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", activityCols, whereClause),
		listArgs...); err != nil {
		return nil, err
	}

	out := make([]*promotion.Activity, 0, len(rows))
	for i := range rows {
		out = append(out, rowToActivityPb(&rows[i]))
	}
	return &promotion.ListActivitiesResp{Activities: out, Total: total}, nil
}
