-- P2 Epic B-5/B-6: shop level system + level upgrade applications
USE mall_shop;

-- Level template (PDD-style: 青铜/白银/黄金/钻石/王者)
CREATE TABLE IF NOT EXISTS `shop_level_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `level` tinyint NOT NULL DEFAULT 1 COMMENT '1=青铜,2=白银,3=黄金,4=钻石,5=王者',
  `name` varchar(32) NOT NULL DEFAULT '',
  `min_gmv` bigint NOT NULL DEFAULT 0 COMMENT 'cents',
  `min_credit_score` int NOT NULL DEFAULT 0,
  `min_months` int NOT NULL DEFAULT 0,
  `min_rating` decimal(3,2) NOT NULL DEFAULT 0.00,
  `commission_rate` decimal(5,4) NOT NULL DEFAULT 0.0000 COMMENT 'platform commission, e.g. 0.0500=5%',
  `traffic_boost` decimal(3,2) NOT NULL DEFAULT 1.00 COMMENT 'traffic multiplier',
  `benefits` varchar(500) NOT NULL DEFAULT '' COMMENT 'JSON array of benefit codes',
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Application records
CREATE TABLE IF NOT EXISTS `shop_level_application` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `current_level` tinyint NOT NULL DEFAULT 1,
  `target_level` tinyint NOT NULL DEFAULT 2,
  `snapshot` varchar(1000) NOT NULL DEFAULT '' COMMENT 'JSON: gmv/credit/months/rating snapshot',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected',
  `admin_id` bigint NOT NULL DEFAULT 0,
  `admin_remark` varchar(500) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed 5 default levels (PDD reference)
INSERT IGNORE INTO `shop_level_template`
  (level, name, min_gmv, min_credit_score, min_months, min_rating, commission_rate, traffic_boost, benefits, create_time)
VALUES
  (1, '青铜', 0,         60,  0,  0.00, 0.0600, 1.00, '["basic"]',                                                UNIX_TIMESTAMP()),
  (2, '白银', 100000,    70,  1,  4.00, 0.0500, 1.20, '["basic","coupon"]',                                       UNIX_TIMESTAMP()),
  (3, '黄金', 1000000,   80,  3,  4.30, 0.0400, 1.50, '["basic","coupon","flash_discount"]',                      UNIX_TIMESTAMP()),
  (4, '钻石', 10000000,  85,  6,  4.60, 0.0300, 2.00, '["basic","coupon","flash_discount","platform_event"]',     UNIX_TIMESTAMP()),
  (5, '王者', 100000000, 90, 12,  4.80, 0.0200, 3.00, '["basic","coupon","flash_discount","platform_event","vip"]', UNIX_TIMESTAMP());
