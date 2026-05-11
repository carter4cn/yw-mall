USE mall_user;

CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `password_hash` varchar(255) NOT NULL DEFAULT '',
  `email` varchar(128) NOT NULL DEFAULT '',
  `role` varchar(32) NOT NULL DEFAULT 'admin' COMMENT 'super_admin/admin',
  `permissions` text COMMENT 'JSON array of permission strings',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=active,0=disabled',
  `create_time` bigint NOT NULL DEFAULT 0,
  `update_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add status column to existing user table for admin user-management
ALTER TABLE `user` ADD COLUMN `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=active,0=disabled';

-- Operation logs (admin/merchant audit trail)
CREATE TABLE IF NOT EXISTS `admin_op_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `actor_id` bigint NOT NULL DEFAULT 0,
  `actor_role` varchar(32) NOT NULL DEFAULT '',
  `method` varchar(16) NOT NULL DEFAULT '',
  `path` varchar(255) NOT NULL DEFAULT '',
  `request_body` mediumtext,
  `status_code` int NOT NULL DEFAULT 0,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_actor` (`actor_id`, `actor_role`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
