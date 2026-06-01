-- S2.3 新人券: coupon_template 加新人专属字段
-- is_new_user_only=1 时 ReceiveCoupon 会校验 user.create_time 在 within_days 内

ALTER TABLE coupon_template
  ADD COLUMN is_new_user_only TINYINT NOT NULL DEFAULT 0
    COMMENT '是否仅新用户可领: 0=全员可领, 1=仅新用户' AFTER per_user_limit,
  ADD COLUMN new_user_within_days INT NOT NULL DEFAULT 7
    COMMENT '新用户判定: 注册后 N 天内算新用户 (is_new_user_only=1 时生效)' AFTER is_new_user_only;
