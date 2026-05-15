-- Sprint 4 — 安全 / 实名 / 合规
-- All tables live in mall_user.

USE mall_user;

-- S4.3 password history (used by both user + admin via subject_type)
CREATE TABLE IF NOT EXISTS `password_history` (
  `id`            bigint unsigned NOT NULL AUTO_INCREMENT,
  `subject_type`  tinyint         NOT NULL COMMENT '1=user 2=admin',
  `subject_id`    bigint unsigned NOT NULL,
  `password_hash` varchar(255)    NOT NULL,
  `create_time`   bigint          NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_subj` (`subject_type`, `subject_id`, `create_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- S4.3 password expiry tracking. MySQL 9.6 doesn't support
-- ADD COLUMN IF NOT EXISTS, so we wrap in a procedure that checks
-- information_schema first → idempotent re-runs.
DROP PROCEDURE IF EXISTS sp_add_last_password_change;
DELIMITER //
CREATE PROCEDURE sp_add_last_password_change()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS
                 WHERE table_schema='mall_user' AND table_name='user'
                   AND column_name='last_password_change') THEN
    ALTER TABLE `user` ADD COLUMN `last_password_change` BIGINT NOT NULL DEFAULT 0;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS
                 WHERE table_schema='mall_user' AND table_name='admin_user'
                   AND column_name='last_password_change') THEN
    ALTER TABLE `admin_user` ADD COLUMN `last_password_change` BIGINT NOT NULL DEFAULT 0;
  END IF;
END//
DELIMITER ;
CALL sp_add_last_password_change();
DROP PROCEDURE sp_add_last_password_change;

-- S4.2 admin IP whitelist
CREATE TABLE IF NOT EXISTS `admin_ip_whitelist` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `admin_id`    bigint unsigned NOT NULL,
  `cidr`        varchar(64)     NOT NULL,
  `note`        varchar(255)    NOT NULL DEFAULT '',
  `create_time` bigint          NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- S4.1 admin MFA
CREATE TABLE IF NOT EXISTS `admin_mfa` (
  `admin_id`         bigint unsigned NOT NULL,
  `totp_secret_enc`  varchar(512)    NOT NULL,
  `backup_codes_enc` varchar(1024)   NOT NULL DEFAULT '',
  `enabled`          tinyint         NOT NULL DEFAULT 0,
  `created_at`       bigint          NOT NULL,
  `last_used_at`     bigint          NOT NULL DEFAULT 0,
  PRIMARY KEY (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- S4.4 KYC
CREATE TABLE IF NOT EXISTS `user_kyc` (
  `user_id`            bigint unsigned NOT NULL,
  `status`             tinyint         NOT NULL DEFAULT 0 COMMENT '0=未提交 1=审核中 2=通过 3=拒绝',
  `real_name_enc`      varchar(512)    NOT NULL DEFAULT '',
  `id_card_no_enc`     varchar(512)    NOT NULL DEFAULT '',
  `id_card_front_url`  varchar(512)    NOT NULL DEFAULT '',
  `id_card_back_url`   varchar(512)    NOT NULL DEFAULT '',
  `face_video_url`     varchar(512)    NOT NULL DEFAULT '',
  `reject_reason`      varchar(255)    NOT NULL DEFAULT '',
  `submit_time`        bigint          NOT NULL DEFAULT 0,
  `audit_time`         bigint          NOT NULL DEFAULT 0,
  `audit_admin_id`     bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`user_id`),
  KEY `idx_status` (`status`, `submit_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
