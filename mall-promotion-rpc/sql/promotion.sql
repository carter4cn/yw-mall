-- 优惠活动模块 Phase 1 DDL
-- Schema: mall_promotion
-- 参考: docs/feat/2026-05-31-promotion-phase1-tech-design.md § 二

-- ============================================================================
-- activity: 活动主体 (满减/折扣/一口价/券 等的统一载体)
-- ============================================================================
CREATE TABLE IF NOT EXISTS activity (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  type            VARCHAR(32)     NOT NULL                    COMMENT '活动类型: fullreduce/discount/fixprice/coupon',
  name            VARCHAR(200)    NOT NULL                    COMMENT '活动名称',
  shop_id         BIGINT UNSIGNED NOT NULL DEFAULT 0          COMMENT '所属店铺, 0=平台活动',
  status          TINYINT         NOT NULL DEFAULT 0          COMMENT '0草稿/1待开始/2进行中/3已结束/4已下线',
  start_time      BIGINT          NOT NULL DEFAULT 0          COMMENT '开始时间 unix',
  end_time        BIGINT          NOT NULL DEFAULT 0          COMMENT '结束时间 unix',
  priority        INT             NOT NULL DEFAULT 0          COMMENT '互斥时数字大优先生效',
  stackable       TINYINT         NOT NULL DEFAULT 1          COMMENT '0互斥/1可叠加',
  description     VARCHAR(500)    NOT NULL DEFAULT ''         COMMENT '活动描述',
  create_user_id  BIGINT UNSIGNED NOT NULL DEFAULT 0          COMMENT '创建人 user_id',
  create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_shop_status  (shop_id, status, end_time),
  INDEX idx_type_status  (type, status),
  INDEX idx_time_range   (start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='促销活动主表';

-- ============================================================================
-- activity_target: 活动适用范围 (一个活动可绑多个 target)
-- ============================================================================
CREATE TABLE IF NOT EXISTS activity_target (
  id           BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id  BIGINT UNSIGNED NOT NULL                       COMMENT 'FK activity.id',
  target_type  VARCHAR(16)     NOT NULL                       COMMENT 'sku/category/shop/user_tag/all',
  target_id    BIGINT UNSIGNED NOT NULL                       COMMENT 'sku_id / category_id / shop_id / tag_id / 0=all',
  create_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_activity (activity_id),
  INDEX idx_lookup   (target_type, target_id)                 COMMENT '价格引擎反查命中活动'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动适用范围';

-- ============================================================================
-- activity_action: 活动优惠动作 (一对多支持阶梯满减)
-- ============================================================================
CREATE TABLE IF NOT EXISTS activity_action (
  id                BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id       BIGINT UNSIGNED NOT NULL                  COMMENT 'FK activity.id',
  action_type       VARCHAR(16)     NOT NULL                  COMMENT 'reduce/discount/cash/fixprice/freeship/gift',
  threshold_type    VARCHAR(8)      NOT NULL DEFAULT 'none'   COMMENT 'none/amount/quantity',
  threshold_value   BIGINT          NOT NULL DEFAULT 0        COMMENT '满 X 元(分) 或 X 件',
  benefit_value     BIGINT          NOT NULL DEFAULT 0        COMMENT '减 Y 元(分) / 打 Y 折(75=7.5折) / 一口价(分)',
  max_discount      BIGINT          NOT NULL DEFAULT 0        COMMENT '折扣类最高优惠上限(分), 0=不限',
  gift_sku_id       BIGINT UNSIGNED NOT NULL DEFAULT 0        COMMENT '赠品 sku',
  step_order        INT             NOT NULL DEFAULT 0        COMMENT '阶梯排序; 引擎按 threshold DESC 取首个达标项',
  create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_activity (activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动优惠动作';

-- ============================================================================
-- coupon_template: 券模板
-- ============================================================================
CREATE TABLE IF NOT EXISTS coupon_template (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id     BIGINT UNSIGNED NOT NULL                    COMMENT 'FK activity.id (券也是 activity 的一种)',
  shop_id         BIGINT UNSIGNED NOT NULL DEFAULT 0          COMMENT '0=平台券',
  name            VARCHAR(200)    NOT NULL,
  type            VARCHAR(16)     NOT NULL                    COMMENT 'full_reduce/discount/cash/freeship',
  value           BIGINT          NOT NULL                    COMMENT '满减面值(分)/折扣(75=7.5折)/立减值(分)',
  min_amount      BIGINT          NOT NULL DEFAULT 0          COMMENT '满 X 元可用(分)',
  max_discount    BIGINT          NOT NULL DEFAULT 0          COMMENT '折扣类最高优惠(分), 0=不限',
  category_id     BIGINT UNSIGNED NOT NULL DEFAULT 0          COMMENT '品类券限定, 0=全品类',
  total_count     INT             NOT NULL                    COMMENT '总发放量',
  received_count  INT             NOT NULL DEFAULT 0          COMMENT '已领数',
  used_count      INT             NOT NULL DEFAULT 0          COMMENT '已使用数',
  per_user_limit  INT             NOT NULL DEFAULT 1          COMMENT '每人限领',
  valid_type      TINYINT         NOT NULL DEFAULT 0          COMMENT '0固定日期 / 1领取后N天',
  valid_days      INT             NOT NULL DEFAULT 0          COMMENT '领取后N天有效',
  valid_start     BIGINT          NOT NULL DEFAULT 0,
  valid_end       BIGINT          NOT NULL DEFAULT 0,
  receive_start   BIGINT          NOT NULL DEFAULT 0,
  receive_end     BIGINT          NOT NULL DEFAULT 0,
  status          TINYINT         NOT NULL DEFAULT 1          COMMENT '0下架/1上架',
  create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_shop_status (shop_id, status, receive_end),
  INDEX idx_activity    (activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='券模板';

-- ============================================================================
-- coupon: 用户已领券
-- ============================================================================
CREATE TABLE IF NOT EXISTS coupon (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  template_id   BIGINT UNSIGNED NOT NULL                      COMMENT 'FK coupon_template.id',
  user_id       BIGINT UNSIGNED NOT NULL,
  shop_id       BIGINT UNSIGNED NOT NULL DEFAULT 0            COMMENT '冗余 shop_id 加速查询',
  status        TINYINT         NOT NULL DEFAULT 0            COMMENT '0未用/1已锁定/2已使用/3已过期',
  order_id      BIGINT UNSIGNED NOT NULL DEFAULT 0            COMMENT '已使用时回填',
  receive_time  BIGINT          NOT NULL,
  expire_time   BIGINT          NOT NULL,
  lock_time     BIGINT          NOT NULL DEFAULT 0            COMMENT '锁定时间, 5min 未支付释放',
  use_time      BIGINT          NOT NULL DEFAULT 0,
  create_time   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_status_template (user_id, template_id, status)  COMMENT '配合 per_user_limit 校验',
  INDEX idx_user_active          (user_id, status, expire_time),
  INDEX idx_order                (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户已领券';
