CREATE DATABASE IF NOT EXISTS mall_logistics CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE mall_logistics;

CREATE TABLE IF NOT EXISTS `shipment` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `tracking_no` VARCHAR(64) NOT NULL,
  `carrier` VARCHAR(32) NOT NULL COMMENT 'kuaidi100 carrier code (sf/jd/zto/...)',
  `status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=created, 1=collected, 2=in_transit, 3=delivering, 4=delivered, 5=exception, 6=returned',
  `subscribe_status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=pending, 1=ok, 2=failed',
  `last_track_time` DATETIME NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_carrier_tracking` (`carrier`, `tracking_no`),
  KEY `idx_order` (`order_id`),
  KEY `idx_user_time` (`user_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `shipment_item` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `shipment_id` BIGINT NOT NULL,
  `order_item_id` BIGINT NOT NULL,
  `product_id` BIGINT NOT NULL,
  `quantity` INT NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_shipment` (`shipment_id`),
  KEY `idx_order_item` (`order_item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `shipment_track` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `shipment_id` BIGINT NOT NULL,
  `track_time` DATETIME NOT NULL,
  `location` VARCHAR(255) NULL,
  `description` VARCHAR(500) NOT NULL,
  `state_kuaidi100` SMALLINT NULL COMMENT 'raw kuaidi100 state code (0..14, 255 for synthetic)',
  `state_internal` TINYINT NOT NULL COMMENT 'mapped internal status 0..6',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shipment_time_desc` (`shipment_id`, `track_time`, `description`(50)),
  KEY `idx_shipment_time` (`shipment_id`, `track_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
