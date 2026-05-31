package engine

import (
	"fmt"
)

// Calculate 主入口 — 见包文档的 5 步流程说明。
func Calculate(items []Item, activities []Activity, coupons []Coupon, shippingFee int64) Result {
	bd := Breakdown{}
	conflicts := []PriceConflict{}

	// ============ Step 1: SKU 级评估 ============
	skuActs := indexSkuActivities(activities)
	for i := range items {
		it := &items[i]
		if it.Quantity <= 0 {
			it.DealPrice = it.OriginalPrice
			continue
		}
		bestPrice := it.OriginalPrice
		bestActID := int64(0)
		for _, act := range skuActs[it.SkuID] {
			for _, a := range act.Actions {
				p := applySkuAction(it.OriginalPrice, a)
				if p < bestPrice {
					bestPrice = p
					bestActID = act.ID
				}
			}
		}
		it.DealPrice = bestPrice
		it.SkuActivityID = bestActID
		if bestActID > 0 {
			bd.SkuLevel = append(bd.SkuLevel, SkuLine{
				SkuID:      it.SkuID,
				ActivityID: bestActID,
				Original:   it.OriginalPrice,
				Deal:       bestPrice,
				Saved:      (it.OriginalPrice - bestPrice) * int64(it.Quantity),
			})
		}
	}

	// ============ Step 2: 店铺级满减 ============
	shopGroups := groupByShop(items)
	shopSubtotal := map[int64]int64{}
	for shopID, group := range shopGroups {
		sub := int64(0)
		totalQty := int32(0)
		for _, it := range group {
			sub += it.DealPrice * int64(it.Quantity)
			totalQty += it.Quantity
		}
		shopSubtotal[shopID] = sub

		shopActs := shopLevelActivitiesFor(activities, shopID)
		bestSaved := int64(0)
		var bestAct *Activity
		for i := range shopActs {
			act := &shopActs[i]
			saved := evaluateShopActivity(act, sub, totalQty)
			if saved > bestSaved {
				bestSaved = saved
				bestAct = act
			}
		}
		if bestAct != nil && bestSaved > 0 {
			bd.ShopLevel = append(bd.ShopLevel, ShopLine{
				ShopID:     shopID,
				ActivityID: bestAct.ID,
				Type:       bestAct.Type,
				Subtotal:   sub,
				Saved:      bestSaved,
			})
			shopSubtotal[shopID] -= bestSaved
		}
	}

	// ============ Step 3: 券应用 ============
	for _, c := range coupons {
		if c.Type == "freeship" {
			continue // step 4 处理
		}
		if c.ShopID > 0 {
			// 店铺券
			sub, ok := shopSubtotal[c.ShopID]
			if !ok {
				conflicts = append(conflicts, PriceConflict{
					CouponID: c.ID, Reason: "购物车无该店铺商品",
				})
				continue
			}
			if sub < c.MinAmount {
				conflicts = append(conflicts, PriceConflict{
					CouponID: c.ID,
					Reason:   fmt.Sprintf("店铺金额未满 %.2f 元门槛", float64(c.MinAmount)/100),
				})
				continue
			}
			saved := applyCoupon(c, sub)
			if saved <= 0 {
				continue
			}
			shopSubtotal[c.ShopID] -= saved
			bd.Coupons = append(bd.Coupons, CouponLine{
				CouponID: c.ID, TemplateID: c.TemplateID, Type: c.Type,
				Scope: "shop", ShopID: c.ShopID, Saved: saved,
			})
		} else {
			// 平台券 (shop_id=0)
			allSub := int64(0)
			for _, v := range shopSubtotal {
				allSub += v
			}
			if allSub < c.MinAmount {
				conflicts = append(conflicts, PriceConflict{
					CouponID: c.ID,
					Reason:   fmt.Sprintf("总金额未满 %.2f 元平台券门槛", float64(c.MinAmount)/100),
				})
				continue
			}
			saved := applyCoupon(c, allSub)
			if saved <= 0 {
				continue
			}
			// 按各店铺占比分摊扣减
			remaining := saved
			var lastShopID int64
			for sid, v := range shopSubtotal {
				if v <= 0 {
					continue
				}
				share := saved * v / allSub
				if shopSubtotal[sid] >= share {
					shopSubtotal[sid] -= share
				} else {
					share = shopSubtotal[sid]
					shopSubtotal[sid] = 0
				}
				remaining -= share
				lastShopID = sid
			}
			if remaining > 0 && lastShopID > 0 {
				// 尾差归到最后一家店铺
				if shopSubtotal[lastShopID] >= remaining {
					shopSubtotal[lastShopID] -= remaining
				}
			}
			bd.Coupons = append(bd.Coupons, CouponLine{
				CouponID: c.ID, TemplateID: c.TemplateID, Type: c.Type,
				Scope: "platform", Saved: saved,
			})
		}
	}

	// ============ Step 4: 运费 / 包邮券 ============
	finalShipping := shippingFee
	for _, c := range coupons {
		if c.Type == "freeship" && shippingFee > 0 {
			bd.ShippingSaved = shippingFee
			finalShipping = 0
			bd.Coupons = append(bd.Coupons, CouponLine{
				CouponID: c.ID, TemplateID: c.TemplateID, Type: c.Type,
				Scope: "shipping", Saved: shippingFee,
			})
			break
		}
	}

	// ============ Step 5: 汇总 ============
	totalOrig := int64(0)
	for _, it := range items {
		if it.Quantity > 0 {
			totalOrig += it.OriginalPrice * int64(it.Quantity)
		}
	}
	promoDiscount := int64(0)
	for _, l := range bd.SkuLevel {
		promoDiscount += l.Saved
	}
	for _, l := range bd.ShopLevel {
		promoDiscount += l.Saved
	}
	couponDiscount := int64(0)
	for _, l := range bd.Coupons {
		if l.Scope != "shipping" {
			couponDiscount += l.Saved
		}
	}
	paid := totalOrig - promoDiscount - couponDiscount + finalShipping
	if paid < 0 {
		paid = 0 // 防呆: 优惠 > 商品额时不能负
	}

	return Result{
		TotalAmount:       totalOrig,
		PromotionDiscount: promoDiscount,
		CouponDiscount:    couponDiscount,
		ShippingFee:       finalShipping,
		PaidAmount:        paid,
		Breakdown:         bd,
		Conflicts:         conflicts,
	}
}
