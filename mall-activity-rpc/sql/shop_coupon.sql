USE mall_activity;
CREATE TABLE IF NOT EXISTS `shop_coupon` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `code` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(128) NOT NULL DEFAULT '',
  `type` tinyint NOT NULL DEFAULT 1 COMMENT '1=满减,2=折扣,3=直减',
  `discount_value` bigint NOT NULL DEFAULT 0 COMMENT 'cents (满减/直减) 或 万分比 (折扣)',
  `min_order_amount` bigint NOT NULL DEFAULT 0,
  `total_quantity` int NOT NULL DEFAULT 0 COMMENT '0=unlimited',
  `claimed_quantity` int NOT NULL DEFAULT 0,
  `per_user_limit` int NOT NULL DEFAULT 1,
  `valid_from` bigint NOT NULL DEFAULT 0,
  `valid_to` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=active,0=disabled',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
