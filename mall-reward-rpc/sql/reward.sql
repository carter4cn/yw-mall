CREATE DATABASE IF NOT EXISTS mall_reward;
USE mall_reward;

CREATE TABLE IF NOT EXISTS `reward_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `type` varchar(32) NOT NULL DEFAULT '' COMMENT 'points/coupon/red_envelope/physical/virtual_goods/member_privilege/experience',
  `payload_schema_json` text,
  `max_value` bigint NOT NULL DEFAULT 0,
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE',
  `description` varchar(255) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `reward_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `activity_id` bigint NOT NULL DEFAULT 0,
  `workflow_instance_id` bigint NOT NULL DEFAULT 0,
  `template_id` bigint NOT NULL,
  `type` varchar(32) NOT NULL DEFAULT '',
  `payload_json` text,
  `status` varchar(32) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/DISPATCHED/CONFIRMED/FAILED/REFUNDED',
  `idempotency_key` varchar(128) NOT NULL DEFAULT '',
  `version` int NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_idem` (`idempotency_key`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_workflow` (`workflow_instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `reward_dispatch_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `reward_record_id` bigint NOT NULL,
  `target_service` varchar(64) NOT NULL DEFAULT '',
  `target_method` varchar(64) NOT NULL DEFAULT '',
  `request` text,
  `response` text,
  `latency_ms` bigint NOT NULL DEFAULT 0,
  `attempt` int NOT NULL DEFAULT 1,
  `success` tinyint NOT NULL DEFAULT 0,
  `error` varchar(1024) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_record` (`reward_record_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `outbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `topic` varchar(128) NOT NULL DEFAULT '',
  `key` varchar(128) NOT NULL DEFAULT '',
  `payload` text,
  `status` varchar(16) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/PUBLISHED/CANCELLED',
  `published_at` timestamp NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_create` (`status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
