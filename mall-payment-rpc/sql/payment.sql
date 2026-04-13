CREATE DATABASE IF NOT EXISTS mall_payment;
USE mall_payment;

CREATE TABLE IF NOT EXISTS `payment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `payment_no` varchar(64) NOT NULL DEFAULT '',
  `order_no` varchar(64) NOT NULL DEFAULT '',
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `amount` bigint NOT NULL DEFAULT 0 COMMENT 'in cents',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=success,2=failed',
  `pay_type` tinyint NOT NULL DEFAULT 0 COMMENT '1=alipay,2=wechat',
  `pay_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_payment_no` (`payment_no`),
  KEY `idx_order_no` (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
