package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

// ChangeActivityStatus 活动状态机：
//   0 草稿 → 1 待开始 (上线)
//   1 待开始 / 2 进行中 → 4 已下线 (下线)
//   2 进行中 → 3 已结束 (由 cron 触发，非用户主动)
// 其它转换全部拒绝。
func ChangeActivityStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ChangeActivityStatusReq) (*promotion.OkResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}

	var cur struct {
		Status    int32 `db:"status"`
		StartTime int64 `db:"start_time"`
		EndTime   int64 `db:"end_time"`
	}
	if err := svcCtx.DB.QueryRowCtx(ctx, &cur,
		"SELECT status, start_time, end_time FROM activity WHERE id = ?", in.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("activity not found")
		}
		return nil, err
	}

	if !validTransition(cur.Status, in.Status) {
		return nil, errors.New("invalid status transition")
	}

	// 上线时根据当前时间自动校准 1 待开始 / 2 进行中
	target := in.Status
	if cur.Status == 0 && in.Status == 1 {
		now := time.Now().Unix()
		if now >= cur.StartTime && now < cur.EndTime {
			target = 2 // 直接到"进行中"
		} else if now >= cur.EndTime {
			return nil, errors.New("activity end_time already passed")
		}
	}

	if _, err := svcCtx.DB.ExecCtx(ctx,
		"UPDATE activity SET status = ? WHERE id = ? AND status = ?",
		target, in.Id, cur.Status,
	); err != nil {
		return nil, err
	}
	return &promotion.OkResp{Ok: true}, nil
}

func validTransition(from, to int32) bool {
	switch from {
	case 0:
		return to == 1 || to == 4
	case 1:
		return to == 2 || to == 4
	case 2:
		return to == 3 || to == 4
	}
	return false
}
