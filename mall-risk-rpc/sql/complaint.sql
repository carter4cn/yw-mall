-- P1 Epic F: complaint & shop restriction
USE mall_risk;

CREATE TABLE IF NOT EXISTS `complaint_ticket` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `complainant_type` varchar(16) NOT NULL DEFAULT '' COMMENT 'user/shop',
  `complainant_id` bigint NOT NULL DEFAULT 0,
  `defendant_type` varchar(16) NOT NULL DEFAULT '' COMMENT 'user/shop',
  `defendant_id` bigint NOT NULL DEFAULT 0,
  `order_id` bigint NOT NULL DEFAULT 0,
  `category` varchar(32) NOT NULL DEFAULT '' COMMENT 'quality/logistics/fraud/service/other',
  `content` varchar(1000) NOT NULL DEFAULT '',
  `evidence_urls` text COMMENT 'JSON array of image URLs',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=open,1=processing,2=closed_support,3=closed_dismiss,4=closed_mediate',
  `admin_id` bigint NOT NULL DEFAULT 0,
  `admin_remark` varchar(500) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_complainant` (`complainant_type`, `complainant_id`),
  KEY `idx_defendant` (`defendant_type`, `defendant_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `shop_restriction` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `restriction` varchar(64) NOT NULL DEFAULT '' COMMENT 'no_new_product/no_activity/no_withdraw',
  `reason` varchar(255) NOT NULL DEFAULT '',
  `operator_id` bigint NOT NULL DEFAULT 0,
  `expire_time` bigint NOT NULL DEFAULT 0 COMMENT '0=permanent',
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
