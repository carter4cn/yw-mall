CREATE DATABASE IF NOT EXISTS mall_user;
USE mall_user;

CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `password` varchar(255) NOT NULL DEFAULT '',
  `phone` varchar(20) NOT NULL DEFAULT '',
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `user_address` (
  `id`            bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`       bigint unsigned NOT NULL DEFAULT 0,
  `receiver_name` varchar(32)     NOT NULL DEFAULT '',
  `phone`         varchar(20)     NOT NULL DEFAULT '',
  `province`      varchar(32)     NOT NULL DEFAULT '',
  `city`          varchar(32)     NOT NULL DEFAULT '',
  `district`      varchar(32)     NOT NULL DEFAULT '',
  `detail`        varchar(255)    NOT NULL DEFAULT '',
  `is_default`    tinyint         NOT NULL DEFAULT 0,
  `create_time`   bigint          NOT NULL DEFAULT 0,
  `update_time`   bigint          NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_user_default` (`user_id`, `is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
