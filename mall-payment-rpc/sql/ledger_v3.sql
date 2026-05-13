USE mall_payment;
CREATE TABLE IF NOT EXISTS `account_ledger` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `direction` tinyint NOT NULL DEFAULT 1 COMMENT '1=credit 入账 2=debit 出账',
  `category` varchar(32) NOT NULL DEFAULT '' COMMENT 'order_income / refund / commission / withdrawal / adjustment',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT 'always positive (cents)',
  `running_balance` bigint NOT NULL DEFAULT 0 COMMENT '此条记录后店家余额',
  `order_id` bigint NOT NULL DEFAULT 0,
  `refund_id` bigint NOT NULL DEFAULT 0,
  `withdrawal_id` bigint NOT NULL DEFAULT 0,
  `ref_no` varchar(64) NOT NULL DEFAULT '' COMMENT 'order_no / refund_no',
  `description` varchar(255) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_time` (`shop_id`, `create_time`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_category_time` (`category`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
