// 价格引擎 TDD 测试矩阵 — 覆盖 § 4.5 全部 16 个 case + 边界。
//
// 跑全套: go test ./internal/engine/...
package engine

import (
	"testing"
)

// helpers ------------------------------------------------------------

func skuItem(skuID, shopID, price int64, qty int32) Item {
	return Item{SkuID: skuID, ShopID: shopID, OriginalPrice: price, Quantity: qty}
}

func fixpriceAct(id, skuID, price int64) Activity {
	return Activity{
		ID: id, Type: "fixprice", ShopID: shopOfSku(skuID),
		Targets: []Target{{TargetType: "sku", TargetID: skuID}},
		Actions: []Action{{ActionType: "fixprice", BenefitValue: price}},
	}
}

func discountSkuAct(id, skuID, percent int64) Activity {
	return Activity{
		ID: id, Type: "discount", ShopID: shopOfSku(skuID),
		Targets: []Target{{TargetType: "sku", TargetID: skuID}},
		Actions: []Action{{ActionType: "discount", BenefitValue: percent}},
	}
}

func shopFullreduceAct(id, shopID int64, steps ...Action) Activity {
	return Activity{
		ID: id, Type: "fullreduce", ShopID: shopID,
		Targets: []Target{{TargetType: "shop", TargetID: shopID}},
		Actions: steps,
	}
}

func step(thr, benefit int64, order int32) Action {
	return Action{
		ActionType: "reduce", ThresholdType: "amount",
		ThresholdValue: thr, BenefitValue: benefit, StepOrder: order,
	}
}

func shopCashCoupon(id, shopID, value, min int64) Coupon {
	return Coupon{
		ID: id, TemplateID: id, ShopID: shopID, Type: "cash",
		Value: value, MinAmount: min,
	}
}

func platformCashCoupon(id, value, min int64) Coupon {
	return Coupon{
		ID: id, TemplateID: id, ShopID: 0, Type: "cash",
		Value: value, MinAmount: min,
	}
}

func freeshipCoupon(id int64) Coupon {
	return Coupon{ID: id, TemplateID: id, Type: "freeship"}
}

// 测试约定: sku 1xx → shop 1, sku 2xx → shop 2
func shopOfSku(sid int64) int64 {
	if sid >= 200 {
		return 2
	}
	return 1
}

// ============================================================================
// TC-01: 单 SKU 无活动
// ============================================================================
func TestTC01_SingleSkuNoActivity(t *testing.T) {
	items := []Item{skuItem(101, 1, 10000, 2)}
	r := Calculate(items, nil, nil, 0)

	if r.TotalAmount != 20000 || r.PaidAmount != 20000 {
		t.Fatalf("TC-01 expected total=20000 paid=20000, got total=%d paid=%d", r.TotalAmount, r.PaidAmount)
	}
	if r.PromotionDiscount != 0 || r.CouponDiscount != 0 {
		t.Fatalf("TC-01 unexpected discounts: %+v", r)
	}
}

// ============================================================================
// TC-02: SKU 命中一口价
// ============================================================================
func TestTC02_SkuFixprice(t *testing.T) {
	items := []Item{skuItem(101, 1, 12900, 1)} // 原价 129
	acts := []Activity{fixpriceAct(5, 101, 9900)} // 一口价 99
	r := Calculate(items, acts, nil, 0)

	if r.PaidAmount != 9900 {
		t.Fatalf("TC-02 expected paid=9900, got %d", r.PaidAmount)
	}
	if r.PromotionDiscount != 3000 {
		t.Fatalf("TC-02 expected promotionDiscount=3000, got %d", r.PromotionDiscount)
	}
	if len(r.Breakdown.SkuLevel) != 1 || r.Breakdown.SkuLevel[0].Deal != 9900 {
		t.Fatalf("TC-02 breakdown wrong: %+v", r.Breakdown.SkuLevel)
	}
}

// ============================================================================
// TC-03: SKU 同时命中一口价 + 7 折，取最优
// ============================================================================
func TestTC03_SkuBestPriceWins(t *testing.T) {
	items := []Item{skuItem(101, 1, 10000, 1)}
	acts := []Activity{
		fixpriceAct(5, 101, 8000),         // 一口价 80
		discountSkuAct(6, 101, 70),        // 7 折 = 70
	}
	r := Calculate(items, acts, nil, 0)

	if r.PaidAmount != 7000 {
		t.Fatalf("TC-03 expected paid=7000 (best of 80/70), got %d", r.PaidAmount)
	}
}

// ============================================================================
// TC-04: 店铺满 199 减 30，购物车 250 元
// ============================================================================
func TestTC04_ShopFullReduce(t *testing.T) {
	items := []Item{skuItem(101, 1, 25000, 1)}
	acts := []Activity{shopFullreduceAct(7, 1, step(19900, 3000, 1))}
	r := Calculate(items, acts, nil, 0)

	if r.PaidAmount != 22000 {
		t.Fatalf("TC-04 expected paid=22000, got %d", r.PaidAmount)
	}
	if r.PromotionDiscount != 3000 {
		t.Fatalf("TC-04 expected promo=3000, got %d", r.PromotionDiscount)
	}
}

// ============================================================================
// TC-05: 店铺阶梯满减，购物车 350 元 → 取 299 减 50 档
// ============================================================================
func TestTC05_StepFullReduce(t *testing.T) {
	items := []Item{skuItem(101, 1, 35000, 1)}
	acts := []Activity{
		shopFullreduceAct(7, 1,
			step(19900, 3000, 1),
			step(29900, 5000, 2),
			step(49900, 10000, 3),
		),
	}
	r := Calculate(items, acts, nil, 0)

	if r.PromotionDiscount != 5000 {
		t.Fatalf("TC-05 expected promo=5000 (299 减 50 档), got %d", r.PromotionDiscount)
	}
}

// ============================================================================
// TC-06: 阶梯门槛差 1 分钱不达标
// ============================================================================
func TestTC06_ThresholdNotMet(t *testing.T) {
	items := []Item{skuItem(101, 1, 19899, 1)} // 198.99 < 199 门槛
	acts := []Activity{shopFullreduceAct(7, 1, step(19900, 3000, 1))}
	r := Calculate(items, acts, nil, 0)

	if r.PromotionDiscount != 0 {
		t.Fatalf("TC-06 expected no discount, got %d", r.PromotionDiscount)
	}
}

// ============================================================================
// TC-07: SKU 级 + 店铺级 + 店铺券三层叠加
// ============================================================================
func TestTC07_ThreeLayerStack(t *testing.T) {
	items := []Item{skuItem(101, 1, 30000, 1)} // 原价 300
	acts := []Activity{
		fixpriceAct(5, 101, 25000),                              // SKU 级降到 250
		shopFullreduceAct(7, 1, step(19900, 3000, 1)),            // 店铺满 199 减 30
	}
	coupons := []Coupon{shopCashCoupon(11, 1, 1000, 5000)}        // 店铺 10 元立减券, min 50
	r := Calculate(items, acts, coupons, 0)

	// 300 → SKU 级 250 → 店铺减 30 = 220 → 券 10 = 210
	if r.PaidAmount != 21000 {
		t.Fatalf("TC-07 expected paid=21000, got %d", r.PaidAmount)
	}
	if r.PromotionDiscount != 8000 {
		t.Fatalf("TC-07 expected promotionDiscount=8000 (sku 50 + shop 30), got %d", r.PromotionDiscount)
	}
	if r.CouponDiscount != 1000 {
		t.Fatalf("TC-07 expected couponDiscount=1000, got %d", r.CouponDiscount)
	}
}

// ============================================================================
// TC-08: 店铺券 + 平台券同时用，平台券基于店铺券后的总额
// ============================================================================
func TestTC08_ShopAndPlatformCoupon(t *testing.T) {
	items := []Item{skuItem(101, 1, 30000, 1)}
	coupons := []Coupon{
		shopCashCoupon(11, 1, 1000, 5000),      // 店铺 -10
		platformCashCoupon(12, 2000, 20000),    // 平台 -20, min 200
	}
	r := Calculate(items, nil, coupons, 0)

	// 300 → 店铺券 -10 = 290 → 平台券 -20 = 270
	if r.PaidAmount != 27000 {
		t.Fatalf("TC-08 expected paid=27000, got %d", r.PaidAmount)
	}
	if r.CouponDiscount != 3000 {
		t.Fatalf("TC-08 expected couponDiscount=3000, got %d", r.CouponDiscount)
	}
}

// ============================================================================
// TC-09: 平台券按店铺占比分摊到各店
// ============================================================================
func TestTC09_PlatformCouponProrata(t *testing.T) {
	items := []Item{
		skuItem(101, 1, 30000, 1), // shop 1: 300
		skuItem(201, 2, 10000, 1), // shop 2: 100
	}
	coupons := []Coupon{platformCashCoupon(12, 4000, 30000)} // 平台 -40, min 300
	r := Calculate(items, nil, coupons, 0)

	// 总 400 → -40 = 360
	if r.PaidAmount != 36000 {
		t.Fatalf("TC-09 expected paid=36000, got %d", r.PaidAmount)
	}
	// 验证 breakdown 里平台券 entry
	found := false
	for _, c := range r.Breakdown.Coupons {
		if c.Scope == "platform" && c.Saved == 4000 {
			found = true
		}
	}
	if !found {
		t.Fatalf("TC-09 platform coupon line missing in breakdown: %+v", r.Breakdown.Coupons)
	}
}

// ============================================================================
// TC-10: 用户传的券不满门槛，返 conflict 不应用
// ============================================================================
func TestTC10_CouponBelowThreshold(t *testing.T) {
	items := []Item{skuItem(101, 1, 3000, 1)} // 30 元
	coupons := []Coupon{shopCashCoupon(11, 1, 1000, 10000)} // 店铺券 min 100
	r := Calculate(items, nil, coupons, 0)

	if r.CouponDiscount != 0 {
		t.Fatalf("TC-10 expected no coupon applied, got %d", r.CouponDiscount)
	}
	if len(r.Conflicts) != 1 || r.Conflicts[0].CouponID != 11 {
		t.Fatalf("TC-10 expected 1 conflict for coupon 11, got %+v", r.Conflicts)
	}
}

// ============================================================================
// TC-11: 折扣券有 max_discount 触发上限
// ============================================================================
func TestTC11_DiscountCouponMaxCap(t *testing.T) {
	items := []Item{skuItem(101, 1, 100000, 1)} // 1000 元
	coupons := []Coupon{{
		ID: 11, TemplateID: 11, ShopID: 1, Type: "discount",
		Value: 70, MinAmount: 0, MaxDiscount: 5000, // 7 折但最多减 50
	}}
	r := Calculate(items, nil, coupons, 0)

	// 7 折应减 300, 但 max=50 截顶
	if r.CouponDiscount != 5000 {
		t.Fatalf("TC-11 expected coupon=5000 (capped), got %d", r.CouponDiscount)
	}
}

// ============================================================================
// TC-12: 多店购物车 + 店铺活动只对 shop1 生效
// ============================================================================
func TestTC12_ShopIsolation(t *testing.T) {
	items := []Item{
		skuItem(101, 1, 25000, 1),
		skuItem(201, 2, 25000, 1),
	}
	acts := []Activity{
		shopFullreduceAct(7, 1, step(19900, 3000, 1)), // 仅 shop 1 满减
	}
	r := Calculate(items, acts, nil, 0)

	if r.PromotionDiscount != 3000 {
		t.Fatalf("TC-12 expected only shop1 减 30, got promo=%d", r.PromotionDiscount)
	}
	if r.PaidAmount != 47000 {
		t.Fatalf("TC-12 expected paid=47000, got %d", r.PaidAmount)
	}
}

// ============================================================================
// TC-13: 包邮券触发，运费归 0
// ============================================================================
func TestTC13_FreeshipCoupon(t *testing.T) {
	items := []Item{skuItem(101, 1, 5000, 1)}
	coupons := []Coupon{freeshipCoupon(15)}
	r := Calculate(items, nil, coupons, 1500) // 运费 15 元

	if r.ShippingFee != 0 {
		t.Fatalf("TC-13 expected shipping=0, got %d", r.ShippingFee)
	}
	if r.PaidAmount != 5000 {
		t.Fatalf("TC-13 expected paid=5000, got %d", r.PaidAmount)
	}
	if r.Breakdown.ShippingSaved != 1500 {
		t.Fatalf("TC-13 expected shippingSaved=1500, got %d", r.Breakdown.ShippingSaved)
	}
}

// ============================================================================
// TC-14: 优惠总额 > 商品总额, paid=0 不能负
// ============================================================================
func TestTC14_PaidNonNegative(t *testing.T) {
	items := []Item{skuItem(101, 1, 1000, 1)} // 10 元
	coupons := []Coupon{shopCashCoupon(11, 1, 5000, 0)} // 50 元立减券, 无门槛
	r := Calculate(items, nil, coupons, 0)

	if r.PaidAmount != 0 {
		t.Fatalf("TC-14 expected paid=0 (clamped non-negative), got %d", r.PaidAmount)
	}
}

// ============================================================================
// TC-15: quantity=0 item 跳过不算入
// ============================================================================
func TestTC15_ZeroQuantityIgnored(t *testing.T) {
	items := []Item{
		skuItem(101, 1, 10000, 0),
		skuItem(102, 1, 5000, 2),
	}
	r := Calculate(items, nil, nil, 0)

	if r.TotalAmount != 10000 {
		t.Fatalf("TC-15 expected total=10000 (only sku 102 counted), got %d", r.TotalAmount)
	}
}

// ============================================================================
// TC-17: 单订单限购 - 超 quota 时活动不应用 + 返 conflict (S2.4)
// ============================================================================
func TestTC17_PerOrderQuotaExceeded(t *testing.T) {
	items := []Item{skuItem(101, 1, 5000, 5)} // 5 件
	acts := []Activity{
		Activity{
			ID: 7, Type: "fullreduce", ShopID: 1,
			Targets: []Target{{TargetType: "shop", TargetID: 1}},
			Actions: []Action{step(10000, 2000, 1)}, // 满 100 减 20
			Rule:    &ActivityRule{PerOrderQuota: 3}, // 单订单最多 3 件
		},
	}
	r := Calculate(items, acts, nil, 0)

	if r.PromotionDiscount != 0 {
		t.Fatalf("TC-17 超 quota 活动不应用, 期望 promo=0, 实际 %d", r.PromotionDiscount)
	}
	if len(r.Conflicts) != 1 || r.Conflicts[0].CouponID != -7 {
		t.Fatalf("TC-17 期望 1 个 activity 级 conflict (id=-7), 实际 %+v", r.Conflicts)
	}
}

// ============================================================================
// TC-18: 限购 quota 未超时活动正常应用 (S2.4)
// ============================================================================
func TestTC18_PerOrderQuotaWithinLimit(t *testing.T) {
	items := []Item{skuItem(101, 1, 5000, 2)} // 2 件
	acts := []Activity{
		Activity{
			ID: 7, Type: "fullreduce", ShopID: 1,
			Targets: []Target{{TargetType: "shop", TargetID: 1}},
			Actions: []Action{step(5000, 2000, 1)},
			Rule:    &ActivityRule{PerOrderQuota: 3},
		},
	}
	r := Calculate(items, acts, nil, 0)
	if r.PromotionDiscount != 2000 {
		t.Fatalf("TC-18 quota 未超应正常减, 期望 2000, 实际 %d", r.PromotionDiscount)
	}
	if len(r.Conflicts) != 0 {
		t.Fatalf("TC-18 不该有 conflict, 实际 %+v", r.Conflicts)
	}
}

// ============================================================================
// TC-16: 同 SKU 不同 fixprice 活动取最低
// ============================================================================
func TestTC16_MultipleFixpriceLowestWins(t *testing.T) {
	items := []Item{skuItem(101, 1, 10000, 1)}
	acts := []Activity{
		fixpriceAct(5, 101, 8000), // 80
		fixpriceAct(6, 101, 6000), // 60 ← 应取这个
	}
	r := Calculate(items, acts, nil, 0)

	if r.PaidAmount != 6000 {
		t.Fatalf("TC-16 expected lowest fixprice 6000, got %d", r.PaidAmount)
	}
}
