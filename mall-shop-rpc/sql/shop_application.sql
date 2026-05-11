USE mall_shop;

ALTER TABLE `shop` ADD COLUMN `owner_user_id` bigint NOT NULL DEFAULT 0 COMMENT 'linked user_id';
ALTER TABLE `shop` ADD COLUMN `credit_score` int NOT NULL DEFAULT 100 COMMENT 'credit score 0-200';
ALTER TABLE `shop` ADD COLUMN `level` tinyint NOT NULL DEFAULT 1 COMMENT '1-5 merchant level';
ALTER TABLE `shop` ADD COLUMN `contact_phone` varchar(20) NOT NULL DEFAULT '';
ALTER TABLE `shop` ADD COLUMN `business_license` varchar(255) NOT NULL DEFAULT '';

CREATE INDEX `idx_owner_user_id` ON `shop` (`owner_user_id`);

CREATE TABLE IF NOT EXISTS `shop_application` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL DEFAULT 0,
  `shop_name` varchar(64) NOT NULL DEFAULT '',
  `logo` varchar(255) NOT NULL DEFAULT '',
  `description` varchar(500) NOT NULL DEFAULT '',
  `contact_phone` varchar(20) NOT NULL DEFAULT '',
  `business_license` varchar(255) NOT NULL DEFAULT '',
  `legal_person` varchar(64) NOT NULL DEFAULT '',
  `id_card_front` varchar(255) NOT NULL DEFAULT '',
  `id_card_back` varchar(255) NOT NULL DEFAULT '',
  `category` varchar(64) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected,3=need_more_info',
  `review_remark` varchar(500) NOT NULL DEFAULT '',
  `reviewer_id` bigint NOT NULL DEFAULT 0,
  `shop_id` bigint NOT NULL DEFAULT 0 COMMENT 'set after approval',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
