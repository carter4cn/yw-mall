package logic

import (
	"context"
	"errors"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CreateActivity 一次性写 activity + activity_target + activity_action 三张表。
// 失败回滚保证 3 张表数据原子一致。新建后状态 = 0 草稿，需要 ChangeActivityStatus 上线。
func CreateActivity(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.CreateActivityReq) (*promotion.CreateActivityResp, error) {
	if in.Type == "" || in.Name == "" {
		return nil, errors.New("type/name required")
	}
	if in.EndTime <= in.StartTime {
		return nil, errors.New("end_time must be after start_time")
	}
	if len(in.Targets) == 0 {
		return nil, errors.New("at least one target required")
	}
	if len(in.Actions) == 0 {
		return nil, errors.New("at least one action required")
	}

	var newId int64
	err := svcCtx.DB.TransactCtx(ctx, func(c context.Context, tx sqlx.Session) error {
		res, err := tx.ExecCtx(c,
			`INSERT INTO activity (type, name, shop_id, status, start_time, end_time, priority, stackable, description, create_user_id)
			 VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?)`,
			in.Type, in.Name, in.ShopId, in.StartTime, in.EndTime, in.Priority, boolToTinyint(in.Stackable), in.Description, in.CreateUserId,
		)
		if err != nil {
			return err
		}
		newId, err = res.LastInsertId()
		if err != nil {
			return err
		}
		for _, t := range in.Targets {
			if _, err := tx.ExecCtx(c,
				`INSERT INTO activity_target (activity_id, target_type, target_id) VALUES (?, ?, ?)`,
				newId, t.TargetType, t.TargetId,
			); err != nil {
				return err
			}
		}
		for _, a := range in.Actions {
			if _, err := tx.ExecCtx(c,
				`INSERT INTO activity_action (activity_id, action_type, threshold_type, threshold_value, benefit_value, max_discount, gift_sku_id, step_order)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				newId, a.ActionType, defaultStr(a.ThresholdType, "none"), a.ThresholdValue, a.BenefitValue, a.MaxDiscount, a.GiftSkuId, a.StepOrder,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &promotion.CreateActivityResp{Id: newId}, nil
}

func boolToTinyint(b bool) int {
	if b {
		return 1
	}
	return 0
}

func defaultStr(s, dft string) string {
	if s == "" {
		return dft
	}
	return s
}
