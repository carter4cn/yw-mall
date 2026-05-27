-- M2 商品多 SKU：一个 product 对应 0..N 个 sku 行。
-- product 表保持单 SKU 兼容字段（price/stock/images = 默认 SKU 聚合视图）。

CREATE TABLE IF NOT EXISTS sku (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  product_id      BIGINT UNSIGNED NOT NULL,
  shop_id         BIGINT UNSIGNED NOT NULL,         -- 冗余，按店铺查 SKU 时用
  sku_code        VARCHAR(64) NOT NULL,             -- 商家自定义编码 / 或 P{pid}-S{idx} 自动生成
  spec_text       VARCHAR(255) NOT NULL DEFAULT '', -- 规格展示文本「黑色/256GB」
  spec_json       VARCHAR(500) NOT NULL DEFAULT '', -- 规格 JSON {"颜色":"黑色","存储":"256GB"}
  price           BIGINT NOT NULL DEFAULT 0,        -- 分
  stock           BIGINT NOT NULL DEFAULT 0,
  image           VARCHAR(500) NOT NULL DEFAULT '', -- SKU 主图（可与 product.images 不同）
  status          TINYINT NOT NULL DEFAULT 1,       -- 1=active, 0=disabled (软删)
  create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_product_status (product_id, status),
  KEY idx_shop (shop_id, status),
  UNIQUE KEY uk_product_code (product_id, sku_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
