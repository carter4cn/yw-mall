USE mall_risk;
CREATE TABLE IF NOT EXISTS `sensitive_word` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `word` varchar(64) NOT NULL DEFAULT '',
  `category` varchar(32) NOT NULL DEFAULT 'general' COMMENT 'spam/porn/violence/political/general',
  `action` varchar(16) NOT NULL DEFAULT 'flag' COMMENT 'flag (mark for review) / block (refuse)',
  `status` tinyint NOT NULL DEFAULT 1,
  `create_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_word` (`word`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
INSERT IGNORE INTO sensitive_word (word, category, action, status, create_time) VALUES
  ('刷单', 'spam', 'flag', 1, UNIX_TIMESTAMP()),
  ('返现', 'spam', 'flag', 1, UNIX_TIMESTAMP()),
  ('色情', 'porn', 'block', 1, UNIX_TIMESTAMP()),
  ('暴力', 'violence', 'flag', 1, UNIX_TIMESTAMP());
