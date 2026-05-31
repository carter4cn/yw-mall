-- Phase 1 优惠活动接入：order 表加 4 列存优惠明细
-- 字段语义见 docs/feat/2026-05-31-promotion-phase1-tech-design.md § 2.7
--
-- 幂等：ALTER 失败 (列已存在) 在 start.sh 里被忽略
ALTER TABLE `order`
  ADD COLUMN promotion_discount BIGINT NOT NULL DEFAULT 0
    COMMENT '活动总优惠(分) = SKU 级 + 店铺级' AFTER total_amount,
  ADD COLUMN coupon_discount    BIGINT NOT NULL DEFAULT 0
    COMMENT '券优惠总额(分)' AFTER promotion_discount,
  ADD COLUMN paid_amount        BIGINT NOT NULL DEFAULT 0
    COMMENT '用户实付(分) = total - promotion - coupon + shipping' AFTER coupon_discount,
  ADD COLUMN discount_detail    JSON   DEFAULT NULL
    COMMENT '优惠明细 JSON, 退款分摊用';
