CREATE DATABASE IF NOT EXISTS mall_product;
USE mall_product;

CREATE TABLE IF NOT EXISTS `category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL DEFAULT '',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0,
  `sort` int NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `product` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '',
  `description` text,
  `price` bigint NOT NULL DEFAULT 0 COMMENT 'price in cents',
  `stock` bigint NOT NULL DEFAULT 0,
  `category_id` bigint unsigned NOT NULL DEFAULT 0,
  `images` varchar(1024) NOT NULL DEFAULT '',
  `shop_id` bigint unsigned NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=on, 0=off',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_shop_status` (`shop_id`, `status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
