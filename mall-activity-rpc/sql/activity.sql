CREATE DATABASE IF NOT EXISTS mall_activity;
USE mall_activity;

CREATE TABLE IF NOT EXISTS `activity` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `title` varchar(255) NOT NULL DEFAULT '',
  `description` text,
  `type` varchar(32) NOT NULL DEFAULT '' COMMENT 'signin/lottery/seckill/coupon',
  `status` varchar(32) NOT NULL DEFAULT 'DRAFT' COMMENT 'DRAFT/PUBLISHED/PAUSED/ENDED',
  `start_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `template_id` bigint NOT NULL DEFAULT 0,
  `rule_set_id` bigint NOT NULL DEFAULT 0,
  `workflow_definition_id` bigint NOT NULL DEFAULT 0,
  `config_json` text,
  `version` int NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`),
  KEY `idx_type_status` (`type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `activity_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `type` varchar(32) NOT NULL DEFAULT '',
  `default_rule_set_id` bigint NOT NULL DEFAULT 0,
  `default_workflow_definition_id` bigint NOT NULL DEFAULT 0,
  `config_json` text,
  `description` varchar(255) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `participation_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `sequence` int NOT NULL DEFAULT 1,
  `workflow_instance_id` bigint NOT NULL DEFAULT 0,
  `status` varchar(32) NOT NULL DEFAULT 'PENDING',
  `payload_json` text,
  `idempotency_key` varchar(128) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_act_user_seq` (`activity_id`, `user_id`, `sequence`),
  UNIQUE KEY `idx_idem` (`idempotency_key`),
  KEY `idx_user_act` (`user_id`, `activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `activity_stat` (
  `activity_id` bigint NOT NULL,
  `participants` bigint NOT NULL DEFAULT 0,
  `winners` bigint NOT NULL DEFAULT 0,
  `stock_left` bigint NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `activity_inventory_snapshot` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint NOT NULL,
  `sku_id` bigint NOT NULL DEFAULT 0,
  `total_stock` bigint NOT NULL DEFAULT 0,
  `current_stock` bigint NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_act_sku` (`activity_id`, `sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
