USE mall_shop;
CREATE TABLE IF NOT EXISTS `shop_lifecycle_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `action` varchar(16) NOT NULL DEFAULT '' COMMENT 'deactivate / pause / resume',
  `reason` varchar(500) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected',
  `admin_id` bigint NOT NULL DEFAULT 0,
  `admin_remark` varchar(500) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
