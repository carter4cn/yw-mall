// Package lua loads embedded Redis Lua scripts at compile time and exposes
// pre-parsed source strings. The activity-rpc hot path uses ScriptRun via
// go-zero's redis client to send these scripts to a Redis Sentinel master.
package lua

import _ "embed"

//go:embed seckill_decr.lua
var SeckillDecr string

//go:embed coupon_claim.lua
var CouponClaim string

//go:embed lottery_pick.lua
var LotteryPick string
