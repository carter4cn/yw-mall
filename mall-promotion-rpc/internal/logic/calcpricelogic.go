// CalcPrice RPC 包装层 —— 把 grpc req 转 engine 输入, 调引擎, 把结果转 grpc resp。
package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"mall-promotion-rpc/internal/engine"
	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

// CalcPrice 拼接活动 + 用户券 → 调引擎 → 序列化 breakdown。
//
// DB 读取流程:
//   1. 拉 cart 涉及的 SKU/Shop 命中的所有 status=2 活动 + 关联 targets/actions
//   2. 拉用户选定的 coupon 行 + 关联 template (校验未过期, 未使用)
//   3. engine.Calculate(...)
//   4. breakdown JSON 序列化进 resp
func CalcPrice(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.CalcPriceReq) (*promotion.CalcPriceResp, error) {
	if len(in.Items) == 0 {
		return nil, errors.New("items required")
	}

	// 收集 sku/shop ids
	skuSet := map[int64]bool{}
	shopSet := map[int64]bool{}
	items := make([]engine.Item, 0, len(in.Items))
	for _, it := range in.Items {
		skuSet[it.SkuId] = true
		shopSet[it.ShopId] = true
		items = append(items, engine.Item{
			SkuID:         it.SkuId,
			ProductID:     it.ProductId,
			ShopID:        it.ShopId,
			CategoryID:    it.CategoryId,
			OriginalPrice: it.OriginalPrice,
			Quantity:      it.Quantity,
		})
	}

	activities, err := loadActiveActivities(ctx, svcCtx, skuSet, shopSet)
	if err != nil {
		return nil, err
	}
	coupons, conflicts, err := loadUserCoupons(ctx, svcCtx, in.UserId, in.CouponIds)
	if err != nil {
		return nil, err
	}

	result := engine.Calculate(items, activities, coupons, in.ShippingFee)
	// 把 logic 层 conflicts (券不存在/不属于用户) 合并到 engine 输出
	result.Conflicts = append(result.Conflicts, conflicts...)

	breakdownJSON, _ := json.Marshal(result.Breakdown)

	pbConflicts := make([]*promotion.PriceConflict, 0, len(result.Conflicts))
	for _, c := range result.Conflicts {
		pbConflicts = append(pbConflicts, &promotion.PriceConflict{
			CouponId: c.CouponID, Reason: c.Reason,
		})
	}

	return &promotion.CalcPriceResp{
		TotalAmount:       result.TotalAmount,
		PromotionDiscount: result.PromotionDiscount,
		CouponDiscount:    result.CouponDiscount,
		ShippingFee:       result.ShippingFee,
		PaidAmount:        result.PaidAmount,
		DiscountDetail:    string(breakdownJSON),
		Conflicts:         pbConflicts,
	}, nil
}

// loadActiveActivities 拉 cart 命中的活动 (status=2 进行中, 时间窗内)。
//
// SQL:
//   SELECT DISTINCT a.* FROM activity a JOIN activity_target t ON t.activity_id=a.id
//   WHERE a.status=2 AND a.end_time > NOW() AND a.start_time <= NOW()
//     AND (
//       (t.target_type='sku'  AND t.target_id IN (...sku ids))
//    OR (t.target_type='shop' AND t.target_id IN (...shop ids))
//    OR (t.target_type='all')
//     )
func loadActiveActivities(ctx context.Context, svcCtx *svc.ServiceContext, skuSet, shopSet map[int64]bool) ([]engine.Activity, error) {
	if len(skuSet) == 0 && len(shopSet) == 0 {
		return nil, nil
	}
	now := time.Now().Unix()

	clauses := []string{}
	args := []any{}
	if len(skuSet) > 0 {
		ids, placeholders := setToInArgs(skuSet)
		clauses = append(clauses, fmt.Sprintf("(t.target_type='sku' AND t.target_id IN (%s))", placeholders))
		args = append(args, ids...)
	}
	if len(shopSet) > 0 {
		ids, placeholders := setToInArgs(shopSet)
		clauses = append(clauses, fmt.Sprintf("(t.target_type='shop' AND t.target_id IN (%s))", placeholders))
		args = append(args, ids...)
	}
	clauses = append(clauses, "t.target_type='all'")
	targetClause := strings.Join(clauses, " OR ")

	q := fmt.Sprintf(`
		SELECT DISTINCT a.id, a.type, a.shop_id, a.priority, a.stackable
		FROM activity a
		JOIN activity_target t ON t.activity_id = a.id
		WHERE a.status = 2 AND a.start_time <= ? AND a.end_time > ?
		  AND (%s)`, targetClause)

	args = append([]any{now, now}, args...)

	type actHead struct {
		Id        int64 `db:"id"`
		Type      string `db:"type"`
		ShopId    int64  `db:"shop_id"`
		Priority  int32  `db:"priority"`
		Stackable int8   `db:"stackable"`
	}
	var heads []actHead
	if err := svcCtx.DB.QueryRowsCtx(ctx, &heads, q, args...); err != nil {
		return nil, err
	}
	if len(heads) == 0 {
		return nil, nil
	}

	// 加载每个活动的 targets + actions
	out := make([]engine.Activity, 0, len(heads))
	for _, h := range heads {
		targets, err := loadActivityTargetsForEngine(ctx, svcCtx, h.Id)
		if err != nil {
			return nil, err
		}
		actions, err := loadActivityActionsForEngine(ctx, svcCtx, h.Id)
		if err != nil {
			return nil, err
		}
		// S2.4 限购规则 (per_order_quota engine 静态校验, per_user/day 后续 S2.4.2)
		var rule *engine.ActivityRule
		pbRule, _ := loadActivityRule(ctx, svcCtx, h.Id)
		if pbRule != nil {
			rule = &engine.ActivityRule{
				PerUserQuota:  pbRule.PerUserQuota,
				PerOrderQuota: pbRule.PerOrderQuota,
				PerDayQuota:   pbRule.PerDayQuota,
			}
		}
		out = append(out, engine.Activity{
			ID: h.Id, Type: h.Type, ShopID: h.ShopId,
			Priority: h.Priority, Stackable: h.Stackable == 1,
			Targets: targets, Actions: actions, Rule: rule,
		})
	}
	return out, nil
}

func loadActivityTargetsForEngine(ctx context.Context, svcCtx *svc.ServiceContext, aid int64) ([]engine.Target, error) {
	var rows []activityTargetRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, activity_id, target_type, target_id FROM activity_target WHERE activity_id = ?", aid); err != nil {
		return nil, err
	}
	out := make([]engine.Target, 0, len(rows))
	for _, r := range rows {
		out = append(out, engine.Target{TargetType: r.TargetType, TargetID: r.TargetId})
	}
	return out, nil
}

func loadActivityActionsForEngine(ctx context.Context, svcCtx *svc.ServiceContext, aid int64) ([]engine.Action, error) {
	var rows []activityActionRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, activity_id, action_type, threshold_type, threshold_value, benefit_value, max_discount, gift_sku_id, step_order FROM activity_action WHERE activity_id = ? ORDER BY step_order, id",
		aid); err != nil {
		return nil, err
	}
	out := make([]engine.Action, 0, len(rows))
	for _, r := range rows {
		out = append(out, engine.Action{
			ActionType: r.ActionType, ThresholdType: r.ThresholdType,
			ThresholdValue: r.ThresholdValue, BenefitValue: r.BenefitValue,
			MaxDiscount: r.MaxDiscount, GiftSkuID: r.GiftSkuId, StepOrder: r.StepOrder,
		})
	}
	return out, nil
}

// loadUserCoupons 拉用户选定的 coupon 行 (校验属于用户 + 未过期 + status=0 未用)。
// 返回 (engine 视角 coupons, conflicts, err)。
func loadUserCoupons(ctx context.Context, svcCtx *svc.ServiceContext, userId int64, couponIds []int64) ([]engine.Coupon, []engine.PriceConflict, error) {
	if len(couponIds) == 0 {
		return nil, nil, nil
	}
	now := time.Now().Unix()

	placeholders := strings.Repeat("?,", len(couponIds))
	placeholders = placeholders[:len(placeholders)-1]
	q := fmt.Sprintf(`
		SELECT c.id, c.template_id, c.user_id, c.status, c.expire_time,
		       t.shop_id, t.type, t.value, t.min_amount, t.max_discount, t.category_id
		FROM coupon c JOIN coupon_template t ON t.id = c.template_id
		WHERE c.id IN (%s)`, placeholders)
	args := make([]any, 0, len(couponIds))
	for _, id := range couponIds {
		args = append(args, id)
	}

	type couponJoinRow struct {
		Id          int64  `db:"id"`
		TemplateId  int64  `db:"template_id"`
		UserId      int64  `db:"user_id"`
		Status      int32  `db:"status"`
		ExpireTime  int64  `db:"expire_time"`
		ShopId      int64  `db:"shop_id"`
		Type        string `db:"type"`
		Value       int64  `db:"value"`
		MinAmount   int64  `db:"min_amount"`
		MaxDiscount int64  `db:"max_discount"`
		CategoryId  int64  `db:"category_id"`
	}
	var rows []couponJoinRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows, q, args...); err != nil {
		return nil, nil, err
	}

	got := map[int64]bool{}
	var out []engine.Coupon
	var conflicts []engine.PriceConflict
	for _, r := range rows {
		got[r.Id] = true
		if r.UserId != userId {
			conflicts = append(conflicts, engine.PriceConflict{CouponID: r.Id, Reason: "不是您的券"})
			continue
		}
		if r.Status != 0 {
			conflicts = append(conflicts, engine.PriceConflict{CouponID: r.Id, Reason: "券状态不可用"})
			continue
		}
		if r.ExpireTime <= now {
			conflicts = append(conflicts, engine.PriceConflict{CouponID: r.Id, Reason: "券已过期"})
			continue
		}
		out = append(out, engine.Coupon{
			ID: r.Id, TemplateID: r.TemplateId, ShopID: r.ShopId,
			Type: r.Type, Value: r.Value, MinAmount: r.MinAmount,
			MaxDiscount: r.MaxDiscount, CategoryID: r.CategoryId,
		})
	}
	// 用户传了但 DB 里不存在的 id
	for _, id := range couponIds {
		if !got[id] {
			conflicts = append(conflicts, engine.PriceConflict{CouponID: id, Reason: "券不存在"})
		}
	}
	return out, conflicts, nil
}

func setToInArgs(s map[int64]bool) ([]any, string) {
	out := make([]any, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	placeholders := strings.Repeat("?,", len(out))
	placeholders = placeholders[:len(placeholders)-1]
	return out, placeholders
}
