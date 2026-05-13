USE mall_order;

ALTER TABLE refund_request ADD COLUMN refund_type tinyint NOT NULL DEFAULT 1 COMMENT '1=refund_only 2=return_refund 3=exchange';
ALTER TABLE refund_request ADD COLUMN return_tracking_no varchar(64) NOT NULL DEFAULT '';
ALTER TABLE refund_request ADD COLUMN return_carrier varchar(64) NOT NULL DEFAULT '';
ALTER TABLE refund_request ADD COLUMN return_ship_time bigint NOT NULL DEFAULT 0 COMMENT '用户寄回时间';
ALTER TABLE refund_request ADD COLUMN return_received_time bigint NOT NULL DEFAULT 0 COMMENT '店家收货时间';
ALTER TABLE refund_request ADD COLUMN return_inspection_passed tinyint NOT NULL DEFAULT 0 COMMENT '0=未验 1=通过 2=拒收';
ALTER TABLE refund_request ADD COLUMN exchange_new_order_id bigint NOT NULL DEFAULT 0 COMMENT '换货生成的新订单 id';
