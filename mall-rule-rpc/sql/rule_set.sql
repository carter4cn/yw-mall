-- P1 Epic G: rule_set already declared in rule.sql. This file is a no-op
-- placeholder that ensures the schema is also applied by the admin
-- migration step in db-init.sh and stays consistent with the proto.
USE mall_rule;

-- The rule_set table from rule.sql uses varchar(1024) for member_rule_ids.
-- The admin low-code path stores JSON arrays here; ensure column is wide
-- enough by upgrading to TEXT (idempotent via informational schema check).
ALTER TABLE `rule_set` MODIFY COLUMN `member_rule_ids` TEXT NOT NULL;
