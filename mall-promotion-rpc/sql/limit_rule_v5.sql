-- S2.4 限购规则 - activity_rule 表 (一对一 with activity)
-- Phase 2 起步: 先支持 per_order_quota (静态校验, 无需 Redis)
-- per_user / per_day quota 后续 S2.4.2 加 Redis 计数器实现

CREATE TABLE IF NOT EXISTS activity_rule (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id     BIGINT UNSIGNED NOT NULL                COMMENT 'FK activity.id, 一对一',
  per_user_quota  INT             NOT NULL DEFAULT 0      COMMENT '每人累计限购, 0=不限',
  per_order_quota INT             NOT NULL DEFAULT 0      COMMENT '单订单限购件数, 0=不限',
  per_day_quota   INT             NOT NULL DEFAULT 0      COMMENT '每人每日限购, 0=不限',
  create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_activity (activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动限购规则';
