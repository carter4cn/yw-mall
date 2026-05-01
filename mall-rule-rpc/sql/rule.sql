CREATE DATABASE IF NOT EXISTS mall_rule;
USE mall_rule;

CREATE TABLE IF NOT EXISTS `rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `description` varchar(255) NOT NULL DEFAULT '',
  `expression` text NOT NULL,
  `lang` varchar(16) NOT NULL DEFAULT 'expr',
  `version` int NOT NULL DEFAULT 1,
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE',
  `json_schema` text,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rule_set` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `op` varchar(8) NOT NULL DEFAULT 'AND' COMMENT 'AND/OR/NOT',
  `member_rule_ids` varchar(1024) NOT NULL DEFAULT '' COMMENT 'JSON array',
  `description` varchar(255) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rule_evaluation_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL,
  `ctx_hash` varchar(64) NOT NULL DEFAULT '',
  `result` tinyint NOT NULL DEFAULT 0,
  `latency_us` bigint NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_rule_time` (`rule_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
