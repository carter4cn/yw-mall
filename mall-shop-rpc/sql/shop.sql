CREATE TABLE `shop` (
  `id`            bigint unsigned NOT NULL AUTO_INCREMENT,
  `name`          varchar(64)  NOT NULL DEFAULT '',
  `logo`          varchar(255) NOT NULL DEFAULT '',
  `banner`        varchar(255) NOT NULL DEFAULT '',
  `description`   varchar(500) NOT NULL DEFAULT '',
  `rating`        decimal(3,2) NOT NULL DEFAULT 5.00,
  `product_count` int          NOT NULL DEFAULT 0,
  `follow_count`  int          NOT NULL DEFAULT 0,
  `status`        tinyint      NOT NULL DEFAULT 1,
  `create_time`   bigint       NOT NULL DEFAULT 0,
  `update_time`   bigint       NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_rating` (`rating`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `shop_follow` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`     bigint unsigned NOT NULL DEFAULT 0,
  `shop_id`     bigint unsigned NOT NULL DEFAULT 0,
  `create_time` bigint          NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_shop` (`user_id`, `shop_id`),
  KEY `idx_shop` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
