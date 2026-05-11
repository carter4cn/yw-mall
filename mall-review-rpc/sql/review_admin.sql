-- P1 Epic E: review admin migrations
USE mall_review;

ALTER TABLE `review` ADD COLUMN `shop_id` bigint NOT NULL DEFAULT 0;
ALTER TABLE `review` ADD INDEX `idx_shop_id` (`shop_id`, `status`, `id`);

CREATE TABLE IF NOT EXISTS `review_delete_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `review_id` bigint NOT NULL DEFAULT 0,
  `shop_id` bigint NOT NULL DEFAULT 0,
  `reason` varchar(500) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected',
  `admin_remark` varchar(500) NOT NULL DEFAULT '',
  `admin_id` bigint NOT NULL DEFAULT 0,
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
