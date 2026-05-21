-- Sprint 5 multi-identifier: user 表加 email + phone/email 双列加密 + blind index.
USE mall_user;

DROP PROCEDURE IF EXISTS sp_user_multi_id_v2;
DELIMITER //
CREATE PROCEDURE sp_user_multi_id_v2()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS
                 WHERE table_schema='mall_user' AND table_name='user'
                   AND column_name='email') THEN
    ALTER TABLE `user`
      ADD COLUMN `email`      VARCHAR(128) NOT NULL DEFAULT '',
      ADD COLUMN `phone_enc`  VARCHAR(512) NOT NULL DEFAULT '',
      ADD COLUMN `email_enc`  VARCHAR(512) NOT NULL DEFAULT '',
      ADD COLUMN `phone_hash` CHAR(64)     NOT NULL DEFAULT '',
      ADD COLUMN `email_hash` CHAR(64)     NOT NULL DEFAULT '',
      ADD INDEX `idx_phone_hash` (`phone_hash`),
      ADD INDEX `idx_email_hash` (`email_hash`);
  END IF;
END//
DELIMITER ;
CALL sp_user_multi_id_v2();
DROP PROCEDURE sp_user_multi_id_v2;
