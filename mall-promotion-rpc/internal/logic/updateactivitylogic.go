package logic

import (
	"context"
	"database/sql"
	"errors"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UpdateActivity 仅允许 status=0 (草稿) 的活动可改。
// targets / actions 全量替换（删旧建新），简单可靠。
func UpdateActivity(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.UpdateActivityReq) (*promotion.OkResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}
	if in.EndTime <= in.StartTime {
		return nil, errors.New("end_time must be after start_time")
	}

	err := svcCtx.DB.TransactCtx(ctx, func(c context.Context, tx sqlx.Session) error {
		// 1) 校验状态 + 加行锁
		var status int32
		if err := tx.QueryRowCtx(c, &status,
			"SELECT status FROM activity WHERE id = ? FOR UPDATE", in.Id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("activity not found")
			}
			return err
		}
		if status != 0 {
			return errors.New("only draft activity can be updated")
		}

		// 2) 更新主表
		if _, err := tx.ExecCtx(c,
			`UPDATE activity SET name = ?, start_time = ?, end_time = ?, priority = ?, stackable = ?, description = ? WHERE id = ?`,
			in.Name, in.StartTime, in.EndTime, in.Priority, boolToTinyint(in.Stackable), in.Description, in.Id,
		); err != nil {
			return err
		}

		// 3) 全量替换 targets
		if _, err := tx.ExecCtx(c, "DELETE FROM activity_target WHERE activity_id = ?", in.Id); err != nil {
			return err
		}
		for _, t := range in.Targets {
			if _, err := tx.ExecCtx(c,
				"INSERT INTO activity_target (activity_id, target_type, target_id) VALUES (?, ?, ?)",
				in.Id, t.TargetType, t.TargetId,
			); err != nil {
				return err
			}
		}

		// 4) 全量替换 actions
		if _, err := tx.ExecCtx(c, "DELETE FROM activity_action WHERE activity_id = ?", in.Id); err != nil {
			return err
		}
		for _, a := range in.Actions {
			if _, err := tx.ExecCtx(c,
				`INSERT INTO activity_action (activity_id, action_type, threshold_type, threshold_value, benefit_value, max_discount, gift_sku_id, step_order)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				in.Id, a.ActionType, defaultStr(a.ThresholdType, "none"), a.ThresholdValue, a.BenefitValue, a.MaxDiscount, a.GiftSkuId, a.StepOrder,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &promotion.OkResp{Ok: true}, nil
}
