-- S1.5 order state-machine timeline columns
USE mall_order;
ALTER TABLE `order` ADD COLUMN `pay_time` bigint NOT NULL DEFAULT 0;
ALTER TABLE `order` ADD COLUMN `ship_time` bigint NOT NULL DEFAULT 0;
ALTER TABLE `order` ADD COLUMN `cancel_time` bigint NOT NULL DEFAULT 0;
ALTER TABLE `order` ADD COLUMN `cancel_reason` varchar(255) NOT NULL DEFAULT '';
