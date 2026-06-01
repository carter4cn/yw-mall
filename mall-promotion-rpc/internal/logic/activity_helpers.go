// 活动相关的共用 DB 读取/序列化函数。
package logic

import (
	"context"
	"time"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

const activityCols = "id, type, name, shop_id, status, start_time, end_time, priority, stackable, description, create_user_id, create_time, update_time"

type activityRow struct {
	Id            int64     `db:"id"`
	Type          string    `db:"type"`
	Name          string    `db:"name"`
	ShopId        int64     `db:"shop_id"`
	Status        int32     `db:"status"`
	StartTime     int64     `db:"start_time"`
	EndTime       int64     `db:"end_time"`
	Priority      int32     `db:"priority"`
	Stackable     int8      `db:"stackable"`
	Description   string    `db:"description"`
	CreateUserId  int64     `db:"create_user_id"`
	CreateTime    time.Time `db:"create_time"`
	UpdateTime    time.Time `db:"update_time"`
}

type activityTargetRow struct {
	Id          int64  `db:"id"`
	ActivityId  int64  `db:"activity_id"`
	TargetType  string `db:"target_type"`
	TargetId    int64  `db:"target_id"`
}

type activityActionRow struct {
	Id              int64  `db:"id"`
	ActivityId      int64  `db:"activity_id"`
	ActionType      string `db:"action_type"`
	ThresholdType   string `db:"threshold_type"`
	ThresholdValue  int64  `db:"threshold_value"`
	BenefitValue    int64  `db:"benefit_value"`
	MaxDiscount     int64  `db:"max_discount"`
	GiftSkuId       int64  `db:"gift_sku_id"`
	StepOrder       int32  `db:"step_order"`
}

func loadActivityRow(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*promotion.Activity, error) {
	var r activityRow
	if err := svcCtx.DB.QueryRowCtx(ctx, &r,
		"SELECT "+activityCols+" FROM activity WHERE id = ? LIMIT 1", id); err != nil {
		return nil, err
	}
	return rowToActivityPb(&r), nil
}

func loadActivityTargets(ctx context.Context, svcCtx *svc.ServiceContext, activityId int64) ([]*promotion.ActivityTarget, error) {
	var rows []activityTargetRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, activity_id, target_type, target_id FROM activity_target WHERE activity_id = ?", activityId); err != nil {
		return nil, err
	}
	out := make([]*promotion.ActivityTarget, 0, len(rows))
	for _, r := range rows {
		out = append(out, &promotion.ActivityTarget{
			Id:         r.Id,
			ActivityId: r.ActivityId,
			TargetType: r.TargetType,
			TargetId:   r.TargetId,
		})
	}
	return out, nil
}

// loadActivityRule 查活动的限购规则。无规则返回 nil (而非 error)。
func loadActivityRule(ctx context.Context, svcCtx *svc.ServiceContext, activityId int64) (*promotion.ActivityRule, error) {
	var r struct {
		PerUserQuota  int32 `db:"per_user_quota"`
		PerOrderQuota int32 `db:"per_order_quota"`
		PerDayQuota   int32 `db:"per_day_quota"`
	}
	err := svcCtx.DB.QueryRowCtx(ctx, &r,
		"SELECT per_user_quota, per_order_quota, per_day_quota FROM activity_rule WHERE activity_id = ?",
		activityId)
	if err != nil {
		// 没规则 = 不限, 当 nil 返回
		return nil, nil
	}
	if r.PerUserQuota == 0 && r.PerOrderQuota == 0 && r.PerDayQuota == 0 {
		return nil, nil // 全 0 也按"不限"对待
	}
	return &promotion.ActivityRule{
		PerUserQuota:  r.PerUserQuota,
		PerOrderQuota: r.PerOrderQuota,
		PerDayQuota:   r.PerDayQuota,
	}, nil
}

func loadActivityActions(ctx context.Context, svcCtx *svc.ServiceContext, activityId int64) ([]*promotion.ActivityAction, error) {
	var rows []activityActionRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, activity_id, action_type, threshold_type, threshold_value, benefit_value, max_discount, gift_sku_id, step_order FROM activity_action WHERE activity_id = ? ORDER BY step_order, id",
		activityId); err != nil {
		return nil, err
	}
	out := make([]*promotion.ActivityAction, 0, len(rows))
	for _, r := range rows {
		out = append(out, &promotion.ActivityAction{
			Id:             r.Id,
			ActivityId:     r.ActivityId,
			ActionType:     r.ActionType,
			ThresholdType:  r.ThresholdType,
			ThresholdValue: r.ThresholdValue,
			BenefitValue:   r.BenefitValue,
			MaxDiscount:    r.MaxDiscount,
			GiftSkuId:      r.GiftSkuId,
			StepOrder:      r.StepOrder,
		})
	}
	return out, nil
}

func rowToActivityPb(r *activityRow) *promotion.Activity {
	return &promotion.Activity{
		Id:           r.Id,
		Type:         r.Type,
		Name:         r.Name,
		ShopId:       r.ShopId,
		Status:       r.Status,
		StartTime:    r.StartTime,
		EndTime:      r.EndTime,
		Priority:     r.Priority,
		Stackable:    r.Stackable == 1,
		Description:  r.Description,
		CreateUserId: r.CreateUserId,
		CreateTime:   r.CreateTime.Unix(),
		UpdateTime:   r.UpdateTime.Unix(),
	}
}
