-- P2 Epic H-2: T+3 settlement support
-- Adds settlement bookkeeping columns to mall_order.order
USE mall_order;

ALTER TABLE `order` ADD COLUMN `complete_time` bigint NOT NULL DEFAULT 0;
ALTER TABLE `order` ADD COLUMN `settle_status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=settled,2=skipped';
ALTER TABLE `order` ADD INDEX `idx_settle` (`settle_status`, `complete_time`);
