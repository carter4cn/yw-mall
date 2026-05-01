CREATE DATABASE IF NOT EXISTS mall_risk;
USE mall_risk;

CREATE TABLE IF NOT EXISTS `blacklist` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `subject_type` varchar(16) NOT NULL DEFAULT '' COMMENT 'user/ip/device',
  `subject_value` varchar(128) NOT NULL DEFAULT '',
  `reason` varchar(255) NOT NULL DEFAULT '',
  `expires_at` timestamp NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_type_value` (`subject_type`, `subject_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rate_limit_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `scope` varchar(128) NOT NULL DEFAULT '',
  `subject_type` varchar(16) NOT NULL DEFAULT '',
  `window_seconds` int NOT NULL DEFAULT 60,
  `max_count` int NOT NULL DEFAULT 100,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_scope_type` (`scope`, `subject_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `risk_score_cache` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `subject_type` varchar(16) NOT NULL DEFAULT '',
  `subject_value` varchar(128) NOT NULL DEFAULT '',
  `score` int NOT NULL DEFAULT 0,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_type_value` (`subject_type`, `subject_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `participation_token` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `jti` varchar(64) NOT NULL DEFAULT '',
  `subject` varchar(128) NOT NULL DEFAULT '',
  `activity_id` bigint NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `used_at` timestamp NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_jti` (`jti`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
