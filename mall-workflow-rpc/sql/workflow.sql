CREATE DATABASE IF NOT EXISTS mall_workflow;
USE mall_workflow;

CREATE TABLE IF NOT EXISTS `workflow_definition` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `description` varchar(255) NOT NULL DEFAULT '',
  `states_json` text NOT NULL,
  `transitions_json` text NOT NULL,
  `version` int NOT NULL DEFAULT 1,
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `workflow_instance` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `definition_id` bigint NOT NULL,
  `activity_id` bigint NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `state` varchar(64) NOT NULL DEFAULT '',
  `payload_json` text,
  `version` int NOT NULL DEFAULT 0,
  `last_event_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_act_user` (`activity_id`, `user_id`),
  KEY `idx_state_def` (`state`, `definition_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `workflow_step_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `instance_id` bigint NOT NULL,
  `from_state` varchar(64) NOT NULL DEFAULT '',
  `to_state` varchar(64) NOT NULL DEFAULT '',
  `trigger` varchar(64) NOT NULL DEFAULT '',
  `latency_ms` bigint NOT NULL DEFAULT 0,
  `error` varchar(1024) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_inst` (`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `asynq_task_archive` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` varchar(128) NOT NULL DEFAULT '',
  `task_type` varchar(64) NOT NULL DEFAULT '',
  `queue` varchar(64) NOT NULL DEFAULT '',
  `payload` text,
  `retried` int NOT NULL DEFAULT 0,
  `terminal_state` varchar(32) NOT NULL DEFAULT '',
  `last_error` varchar(1024) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`task_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
