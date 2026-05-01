-- seckill_decr.lua
-- KEYS[1] = activity stock hash, e.g. "activity:1001:stock"
-- KEYS[2] = participation dedup set, e.g. "activity:1001:dedup"
-- ARGV[1] = sku_id (string)
-- ARGV[2] = user_id (string)
-- ARGV[3] = quantity to decrement (string, positive integer)
-- Returns:
--   stock_left (>= 0)  on success
--  -1                  if user already participated (dedup)
--  -2                  if sku not found
--  -3                  if stock empty / would go negative

local stock_key = KEYS[1]
local dedup_key = KEYS[2]
local sku_id    = ARGV[1]
local user_id   = ARGV[2]
local qty       = tonumber(ARGV[3])

-- 1. dedup check (one-shot per user/activity)
if redis.call('SADD', dedup_key, user_id) == 0 then
  return -1
end

-- 2. ensure sku exists
local cur = redis.call('HGET', stock_key, sku_id)
if cur == false then
  -- rollback dedup so user can retry against another sku
  redis.call('SREM', dedup_key, user_id)
  return -2
end

-- 3. decrement if enough
if tonumber(cur) < qty then
  redis.call('SREM', dedup_key, user_id)
  return -3
end

local left = redis.call('HINCRBY', stock_key, sku_id, -qty)
if left < 0 then
  -- race lost; restore both
  redis.call('HINCRBY', stock_key, sku_id, qty)
  redis.call('SREM', dedup_key, user_id)
  return -3
end
return left
