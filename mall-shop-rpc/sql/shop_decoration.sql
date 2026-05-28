-- M6 店铺装修配置
-- 一店一行（UNIQUE shop_id），存简单的 banner/announcement/featured_pids
-- 复杂模块化装修后续 Sprint 再加（design spec 5.1）

CREATE TABLE IF NOT EXISTS shop_decoration (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  shop_id       BIGINT UNSIGNED NOT NULL UNIQUE,
  banners       TEXT NOT NULL,                       -- JSON array: [{image,link,sort}]
                                                     -- 或简化为 url csv: "url1,url2,..."
  announcement  VARCHAR(500) NOT NULL DEFAULT '',
  featured_pids VARCHAR(500) NOT NULL DEFAULT '',    -- pid csv: "81,79,77"，最多 10 个
  update_time   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
