// Package lua embeds the Redis Lua scripts the risk service runs.
package lua

import _ "embed"

//go:embed sliding_window.lua
var SlidingWindow string
