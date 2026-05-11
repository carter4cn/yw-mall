USE mall_activity;
CREATE TABLE IF NOT EXISTS `sku_flash_discount` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `product_id` bigint NOT NULL DEFAULT 0,
  `sku_id` bigint NOT NULL DEFAULT 0,
  `original_price` bigint NOT NULL DEFAULT 0 COMMENT 'cents',
  `discount_price` bigint NOT NULL DEFAULT 0 COMMENT 'cents',
  `start_time` bigint NOT NULL DEFAULT 0,
  `end_time` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=active,2=cancelled,3=ended',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop` (`shop_id`, `status`),
  KEY `idx_sku_active` (`sku_id`, `status`, `end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
