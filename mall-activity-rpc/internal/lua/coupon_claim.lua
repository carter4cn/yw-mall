-- coupon_claim.lua
-- KEYS[1] = activity stock counter key, e.g. "activity:2001:coupon_stock"
-- KEYS[2] = per-user counter hash,    e.g. "activity:2001:user_claims"
-- ARGV[1] = user_id (string)
-- ARGV[2] = max_per_user (integer)
-- Returns:
--   stock_left (>= 0)   on success
--  -1                    if user reached per-user limit
--  -2                    if global stock empty

local stock_key   = KEYS[1]
local user_key    = KEYS[2]
local user_id     = ARGV[1]
local max_per_user = tonumber(ARGV[2])

local user_claims = tonumber(redis.call('HGET', user_key, user_id) or '0')
if user_claims >= max_per_user then
  return -1
end

local stock = tonumber(redis.call('GET', stock_key) or '0')
if stock <= 0 then
  return -2
end

local left = redis.call('DECR', stock_key)
if left < 0 then
  -- race lost; restore
  redis.call('INCR', stock_key)
  return -2
end
redis.call('HINCRBY', user_key, user_id, 1)
return left
