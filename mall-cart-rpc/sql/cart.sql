CREATE DATABASE IF NOT EXISTS mall_cart;
USE mall_cart;

CREATE TABLE IF NOT EXISTS `cart_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `product_id` bigint unsigned NOT NULL DEFAULT 0,
  `quantity` int NOT NULL DEFAULT 0,
  `selected` tinyint(1) NOT NULL DEFAULT 1,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_product` (`user_id`, `product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
