USE mall_product;

ALTER TABLE `product` ADD COLUMN `review_status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=approved,2=rejected';
ALTER TABLE `product` ADD COLUMN `review_remark` varchar(500) NOT NULL DEFAULT '';
ALTER TABLE `product` ADD COLUMN `detail` mediumtext COMMENT 'rich text detail HTML';
ALTER TABLE `product` ADD COLUMN `brand` varchar(64) NOT NULL DEFAULT '';
ALTER TABLE `product` ADD COLUMN `weight` decimal(10,2) NOT NULL DEFAULT 0;

CREATE INDEX `idx_review_status` ON `product` (`review_status`, `id`);
