-- lottery_pick.lua
-- Atomically picks a prize index using cumulative-weight sampling and
-- decrements its remaining quota. The prize pool is stored as a single
-- list "activity:N:prizes" of JSON-encoded {weight, remaining, prize_id}
-- entries. ARGV[1] is a random number in [0, total_weight) chosen by the
-- caller (so the Lua script stays deterministic given that input).
--
-- KEYS[1] = prize pool key
-- ARGV[1] = random integer in [0, total_weight)
-- Returns: index (>=0) of the prize that was decremented, or -1 if all
-- remaining quotas are zero.

local key = KEYS[1]
local rand = tonumber(ARGV[1])

local n = tonumber(redis.call('LLEN', key) or '0')
if n == 0 then return -1 end

local cum = 0
local pickedIndex = -1
for i = 0, n - 1 do
  local raw = redis.call('LINDEX', key, i)
  if raw then
    local _, _, w_s, r_s = string.find(raw, '"weight":(%-?%d+),"remaining":(%-?%d+)')
    local w = tonumber(w_s) or 0
    local r = tonumber(r_s) or 0
    cum = cum + w
    if pickedIndex == -1 and rand < cum and r > 0 then
      pickedIndex = i
      -- decrement remaining: rewrite the entry
      local newRaw = string.gsub(raw, '"remaining":' .. r_s, '"remaining":' .. tostring(r - 1))
      redis.call('LSET', key, i, newRaw)
      return i
    end
  end
end

-- all remaining quotas are zero, fall back: any non-zero entry
for i = 0, n - 1 do
  local raw = redis.call('LINDEX', key, i)
  if raw then
    local _, _, _, r_s = string.find(raw, '"weight":(%-?%d+),"remaining":(%-?%d+)')
    local r = tonumber(r_s) or 0
    if r > 0 then
      local newRaw = string.gsub(raw, '"remaining":' .. r_s, '"remaining":' .. tostring(r - 1))
      redis.call('LSET', key, i, newRaw)
      return i
    end
  end
end
return -1
