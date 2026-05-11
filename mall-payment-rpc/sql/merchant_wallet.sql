-- P1 Epic H: merchant wallet, withdrawal, bills
USE mall_payment;

CREATE TABLE IF NOT EXISTS `merchant_wallet` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `balance` bigint NOT NULL DEFAULT 0 COMMENT 'available in cents',
  `frozen` bigint NOT NULL DEFAULT 0 COMMENT 'frozen for refund reserve',
  `total_income` bigint NOT NULL DEFAULT 0,
  `total_withdrawn` bigint NOT NULL DEFAULT 0,
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `withdrawal_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `amount` bigint NOT NULL DEFAULT 0 COMMENT 'in cents',
  `bank_info` varchar(500) NOT NULL DEFAULT '' COMMENT 'JSON: bank/account/name',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected,3=paid',
  `admin_id` bigint NOT NULL DEFAULT 0,
  `admin_remark` varchar(255) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `bill_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `type` varchar(32) NOT NULL DEFAULT '' COMMENT 'income/refund/withdrawal/fee',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT 'positive=income, negative=deduction',
  `order_id` bigint NOT NULL DEFAULT 0,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id_time` (`shop_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
