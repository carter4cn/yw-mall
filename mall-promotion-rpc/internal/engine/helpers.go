package engine

// indexSkuActivities 把活动按 SKU id 索引，便于 Step 1 反查。
//
// 命中规则:
//   - target_type=sku 且 target_id=item.SkuID → 命中
//   - target_type=all (target_id=0) 不在 SKU 级评估里, 走店铺级
// 仅返回 type 为 sku 级活动 (fixprice/discount with sku target)。
func indexSkuActivities(activities []Activity) map[int64][]Activity {
	idx := map[int64][]Activity{}
	for _, act := range activities {
		// SKU 级活动: type 是 fixprice 或 discount, 且有 sku target
		if act.Type != "fixprice" && act.Type != "discount" {
			continue
		}
		for _, t := range act.Targets {
			if t.TargetType == "sku" && t.TargetID > 0 {
				idx[t.TargetID] = append(idx[t.TargetID], act)
			}
		}
	}
	return idx
}

// shopLevelActivitiesFor 返回适用于指定 shop 的店铺级活动 (满减/折扣 with shop target)。
func shopLevelActivitiesFor(activities []Activity, shopID int64) []Activity {
	out := []Activity{}
	for _, act := range activities {
		// 店铺级常见 type: fullreduce, discount (with shop target)
		if act.Type != "fullreduce" && act.Type != "discount" {
			continue
		}
		// shop_id 隔离: 平台活动 (shop_id=0) 暂不在 phase 1 范围
		if act.ShopID != shopID {
			continue
		}
		// target 必须包含 shop_id 或 all
		hit := false
		for _, t := range act.Targets {
			if t.TargetType == "shop" && t.TargetID == shopID {
				hit = true
				break
			}
			if t.TargetType == "all" {
				hit = true
				break
			}
		}
		if hit {
			out = append(out, act)
		}
	}
	return out
}

// groupByShop 按 shop_id 把 items 分组 (返回指针, 便于在外层修改 DealPrice)。
func groupByShop(items []Item) map[int64][]*Item {
	out := map[int64][]*Item{}
	for i := range items {
		it := &items[i]
		if it.Quantity <= 0 {
			continue
		}
		out[it.ShopID] = append(out[it.ShopID], it)
	}
	return out
}

// applySkuAction SKU 级活动评估，返回新单价。
//
//   fixprice: 直接覆盖为 BenefitValue
//   discount: 折扣率 BenefitValue (75 表示 7.5 折)，再用 max_discount 截顶
//   reduce  : 满足 threshold 后减 BenefitValue (SKU 级少见)
func applySkuAction(orig int64, a Action) int64 {
	switch a.ActionType {
	case "fixprice":
		if a.BenefitValue >= 0 && a.BenefitValue < orig {
			return a.BenefitValue
		}
		return orig
	case "discount":
		if a.BenefitValue <= 0 || a.BenefitValue >= 100 {
			return orig
		}
		newPrice := orig * a.BenefitValue / 100
		if a.MaxDiscount > 0 && (orig-newPrice) > a.MaxDiscount {
			newPrice = orig - a.MaxDiscount
		}
		return newPrice
	case "reduce":
		if a.ThresholdType == "amount" && orig < a.ThresholdValue {
			return orig
		}
		if orig <= a.BenefitValue {
			return 0
		}
		return orig - a.BenefitValue
	}
	return orig
}

// evaluateShopActivity 店铺级满减/满折评估, 返回省了多少 (分)。
//
// 阶梯满减: 取 ThresholdValue 最大且 subtotal/qty 仍达标的那档。
func evaluateShopActivity(act *Activity, subtotal int64, totalQty int32) int64 {
	var best *Action
	for i := range act.Actions {
		a := &act.Actions[i]
		hit := false
		switch a.ThresholdType {
		case "amount":
			hit = subtotal >= a.ThresholdValue
		case "quantity":
			hit = int64(totalQty) >= a.ThresholdValue
		case "none", "":
			hit = true
		}
		if !hit {
			continue
		}
		if best == nil || a.ThresholdValue > best.ThresholdValue {
			best = a
		}
	}
	if best == nil {
		return 0
	}
	switch best.ActionType {
	case "reduce":
		return best.BenefitValue
	case "discount":
		saved := subtotal - subtotal*best.BenefitValue/100
		if best.MaxDiscount > 0 && saved > best.MaxDiscount {
			saved = best.MaxDiscount
		}
		return saved
	}
	return 0
}

// applyCoupon 单张券能省多少 (调用方已校验 min_amount)。
func applyCoupon(c Coupon, basis int64) int64 {
	switch c.Type {
	case "cash":
		if c.Value > basis {
			return basis
		}
		return c.Value
	case "full_reduce":
		if c.Value > basis {
			return basis
		}
		return c.Value
	case "discount":
		if c.Value <= 0 || c.Value >= 100 {
			return 0
		}
		saved := basis - basis*c.Value/100
		if c.MaxDiscount > 0 && saved > c.MaxDiscount {
			saved = c.MaxDiscount
		}
		return saved
	}
	return 0
}
