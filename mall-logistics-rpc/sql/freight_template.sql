USE mall_logistics;
CREATE TABLE IF NOT EXISTS `freight_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `name` varchar(128) NOT NULL DEFAULT '',
  `calc_type` tinyint NOT NULL DEFAULT 1 COMMENT '1=件数,2=重量',
  `first_value` int NOT NULL DEFAULT 1 COMMENT '首件件数 / 首件重量(克)',
  `first_fee` bigint NOT NULL DEFAULT 0 COMMENT 'cents',
  `extra_value` int NOT NULL DEFAULT 1,
  `extra_fee` bigint NOT NULL DEFAULT 0,
  `regions` varchar(1000) NOT NULL DEFAULT '' COMMENT 'JSON: 适用省份码列表，空=全国',
  `is_default` tinyint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1,
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
