USE mall_order;

ALTER TABLE `order` ADD COLUMN `shop_id` bigint NOT NULL DEFAULT 0;
ALTER TABLE `order` ADD COLUMN `refund_status` tinyint NOT NULL DEFAULT 0 COMMENT '0=none,1=requested,2=approved,3=rejected,4=completed';
ALTER TABLE `order` ADD COLUMN `refund_reason` varchar(500) NOT NULL DEFAULT '';

CREATE INDEX `idx_shop_status` ON `order` (`shop_id`, `status`, `id`);
