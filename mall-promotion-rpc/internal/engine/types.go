// Package engine 价格计算纯函数引擎。
//
// 设计原则:
//   1. 纯函数 — 所有 DB IO 在 logic 层完成, engine 只接收已就绪数据
//   2. 独立类型 — 不依赖 protobuf, 让单测脱离 grpc 语境
//   3. 顺序确定性 — 固定 5 步流程, 每步产出可解释的 breakdown
//   4. 失败兜底 — 单步 (如某券不可用) 不阻塞整体计算
//
// 5 步流程:
//   1) SKU 级活动评估 (fixprice / discount → 单价覆盖)
//   2) 店铺级满减 (按 shop 分组, 取最优阶梯)
//   3) 券应用 (店铺券 → 平台券, 平台券按占比分摊到各店)
//   4) 运费 / 包邮券
//   5) 汇总: total - promotion - coupon + shipping = paid (>=0 防呆)
package engine

// Item 购物车单项。
type Item struct {
	SkuID         int64
	ProductID     int64
	ShopID        int64
	CategoryID    int64
	OriginalPrice int64 // 单价(分)
	Quantity      int32

	// 评估期间填充
	DealPrice     int64 // 经 SKU 级活动后的单价
	SkuActivityID int64 // 命中的 SKU 级活动 id
}

// Activity 活动 (engine 视角, 不带 db row 字段)。
type Activity struct {
	ID        int64
	Type      string // fullreduce/discount/fixprice/coupon
	ShopID    int64
	Priority  int32
	Stackable bool

	Targets []Target
	Actions []Action
	Rule    *ActivityRule // S2.4 限购, nil 表示不限
}

// ActivityRule 限购规则 (engine 视角)。
type ActivityRule struct {
	PerUserQuota  int32 // 累计 (需 logic 层调用 DB 校验)
	PerOrderQuota int32 // 单订单限购 (engine 静态校验)
	PerDayQuota   int32 // 每人每日 (需 Redis 计数, 暂留 S2.4.2)
}

type Target struct {
	TargetType string // sku/category/shop/all
	TargetID   int64
}

type Action struct {
	ActionType     string // reduce/discount/cash/fixprice/freeship/gift
	ThresholdType  string // none/amount/quantity
	ThresholdValue int64
	BenefitValue   int64
	MaxDiscount    int64
	GiftSkuID      int64
	StepOrder      int32
}

// Coupon 用户已选定的券 (logic 层已校验 status=未用、未过期)。
type Coupon struct {
	ID         int64
	TemplateID int64
	ShopID     int64 // 0 = 平台券
	Type       string // full_reduce/discount/cash/freeship
	Value      int64
	MinAmount  int64
	MaxDiscount int64
	CategoryID int64 // 0 = 全品类
}

// PriceConflict 用户传的某张券不能应用时的说明。
type PriceConflict struct {
	CouponID int64
	Reason   string
}

// SkuLine SKU 级优惠明细。
type SkuLine struct {
	SkuID      int64 `json:"skuId"`
	ActivityID int64 `json:"activityId"`
	Original   int64 `json:"original"`
	Deal       int64 `json:"deal"`
	Saved      int64 `json:"saved"`
}

// ShopLine 店铺级优惠明细。
type ShopLine struct {
	ShopID     int64  `json:"shopId"`
	ActivityID int64  `json:"activityId"`
	Type       string `json:"type"`
	Subtotal   int64  `json:"subtotal"`
	Saved      int64  `json:"saved"`
}

// CouponLine 券优惠明细。
type CouponLine struct {
	CouponID   int64  `json:"couponId"`
	TemplateID int64  `json:"templateId"`
	Type       string `json:"type"`
	Scope      string `json:"scope"` // shop/platform
	ShopID     int64  `json:"shopId,omitempty"`
	Saved      int64  `json:"saved"`
}

// Breakdown 完整优惠明细 — 序列化为 JSON 存到 order.discount_detail 列。
type Breakdown struct {
	SkuLevel      []SkuLine    `json:"sku_level,omitempty"`
	ShopLevel     []ShopLine   `json:"shop_level,omitempty"`
	Coupons       []CouponLine `json:"coupons,omitempty"`
	ShippingSaved int64        `json:"shipping_saved,omitempty"`
}

// Result 引擎最终输出。
type Result struct {
	TotalAmount       int64           // 商品原价合计
	PromotionDiscount int64           // SKU + 店铺级活动总优惠
	CouponDiscount    int64           // 券总优惠
	ShippingFee       int64           // 实际运费 (可能被包邮券减为 0)
	PaidAmount        int64           // 应付 = total - promotion - coupon + shipping (>=0 防呆)
	Breakdown         Breakdown       // 结构化, 序列化存 order
	Conflicts         []PriceConflict // 用户传的不可用券说明
}
