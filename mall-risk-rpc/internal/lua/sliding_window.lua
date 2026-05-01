-- KEYS[1] = sorted-set key (one per scope)
-- ARGV[1] = current epoch ms
-- ARGV[2] = window milliseconds
-- ARGV[3] = max count
-- Returns: {allowed=1|0, remaining, reset_at_ms}
--
-- Sliding-window rate limiter implemented as a sorted set keyed by epoch_ms.
-- Each request inserts (member=epoch_ms+rand, score=epoch_ms). Old entries
-- are GC'd by ZREMRANGEBYSCORE. Cardinality of the trimmed set is the count.
local key = KEYS[1]
local now_ms = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local maxn = tonumber(ARGV[3])

local cutoff = now_ms - window_ms
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)
local cnt = tonumber(redis.call('ZCARD', key))

if cnt >= maxn then
    -- earliest entry's expiry is the soonest we can let one in again
    local earliest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local reset_at = now_ms + window_ms
    if earliest and earliest[2] then
        reset_at = tonumber(earliest[2]) + window_ms
    end
    return {0, 0, reset_at}
end

-- member must be unique → suffix with current count
redis.call('ZADD', key, now_ms, now_ms .. ':' .. cnt)
redis.call('PEXPIRE', key, window_ms)
return {1, maxn - cnt - 1, now_ms + window_ms}
