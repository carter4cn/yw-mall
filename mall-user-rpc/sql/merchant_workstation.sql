-- M1 商家工作台：员工 RBAC + 邀请链接
-- 表归属 mall_user 库（与 user 表同库，因 FK user.id）

CREATE TABLE IF NOT EXISTS merchant_staff (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  shop_id         BIGINT UNSIGNED NOT NULL,
  user_id         BIGINT UNSIGNED NOT NULL,
  role            VARCHAR(32) NOT NULL,            -- owner|service|warehouse|finance
  status          TINYINT NOT NULL DEFAULT 1,      -- 1=active, 0=disabled
  invited_by      BIGINT UNSIGNED NOT NULL DEFAULT 0,
  joined_at       BIGINT NOT NULL,
  create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_shop_user (shop_id, user_id),
  KEY idx_user (user_id),
  KEY idx_shop_status (shop_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS merchant_staff_invitation (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  shop_id         BIGINT UNSIGNED NOT NULL,
  invited_by      BIGINT UNSIGNED NOT NULL,        -- staff.user_id（店主 uid）
  target_phone    VARCHAR(20) NOT NULL DEFAULT '',
  target_email    VARCHAR(255) NOT NULL DEFAULT '',
  role            VARCHAR(32) NOT NULL,
  invitation_code VARCHAR(64) NOT NULL UNIQUE,
  status          TINYINT NOT NULL DEFAULT 0,      -- 0=pending, 1=accepted, 2=expired, 3=revoked
  expires_at      BIGINT NOT NULL,                 -- unix +7d
  accepted_by     BIGINT UNSIGNED NOT NULL DEFAULT 0,
  accepted_at     BIGINT NOT NULL DEFAULT 0,
  create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_shop (shop_id),
  KEY idx_status (status, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
