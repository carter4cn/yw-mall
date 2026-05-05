CREATE TABLE shop (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name          VARCHAR(64)  NOT NULL,
  logo          VARCHAR(255) NOT NULL DEFAULT '',
  banner        VARCHAR(255) NOT NULL DEFAULT '',
  description   VARCHAR(500) NOT NULL DEFAULT '',
  rating        DECIMAL(3,2) NOT NULL DEFAULT 5.00,
  product_count INT          NOT NULL DEFAULT 0,
  follow_count  INT          NOT NULL DEFAULT 0,
  status        TINYINT      NOT NULL DEFAULT 1,
  create_time   BIGINT       NOT NULL,
  update_time   BIGINT       NOT NULL,
  KEY idx_status (status),
  KEY idx_rating (rating DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE shop_follow (
  id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id     BIGINT UNSIGNED NOT NULL,
  shop_id     BIGINT UNSIGNED NOT NULL,
  create_time BIGINT          NOT NULL,
  UNIQUE KEY uk_user_shop (user_id, shop_id),
  KEY idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
