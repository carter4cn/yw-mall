CREATE DATABASE IF NOT EXISTS mall_review CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE mall_review;

CREATE TABLE IF NOT EXISTS `review` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `order_item_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `product_id` BIGINT NOT NULL,
  `score_overall` TINYINT NOT NULL,
  `score_match` TINYINT NOT NULL,
  `score_logistics` TINYINT NOT NULL,
  `score_service` TINYINT NOT NULL,
  `content` VARCHAR(2000) NOT NULL,
  `has_media` TINYINT NOT NULL DEFAULT 0,
  `followup_content` VARCHAR(500) NULL,
  `followup_time` DATETIME NULL,
  `merchant_reply_text` VARCHAR(500) NULL,
  `merchant_reply_time` DATETIME NULL,
  `merchant_user_id` BIGINT NULL,
  `status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=normal, 1=admin_soft_deleted, 2=admin_hidden',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_item` (`order_item_id`),
  KEY `idx_product_status_time` (`product_id`, `status`, `create_time`),
  KEY `idx_product_score` (`product_id`, `score_overall`, `status`),
  KEY `idx_user_time` (`user_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `review_media` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `review_id` BIGINT NOT NULL,
  `media_type` TINYINT NOT NULL COMMENT '1=image, 2=video',
  `media_url` VARCHAR(500) NOT NULL,
  `sort` TINYINT NOT NULL DEFAULT 0,
  `is_followup` TINYINT NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_review` (`review_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
