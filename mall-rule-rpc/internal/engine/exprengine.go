package engine

import (
	"fmt"
	"time"

	"mall-rule-rpc/rule"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Compile parses the given expression with the standard RuleContext env
// and returns a runnable program. Errors include both parse and type errors.
func Compile(expression string) (*vm.Program, error) {
	return expr.Compile(
		expression,
		expr.Env(envType()),
		expr.AsBool(),
		expr.AllowUndefinedVariables(),
		expr.Function("now", funcNow, new(func() int64)),
		expr.Function("duration", funcDuration, new(func(string) int64)),
		expr.Function("inSet", funcInSet, new(func(string, []string) bool)),
	)
}

// Run evaluates a compiled program against a RuleContext.
func Run(p *vm.Program, ctx *rule.RuleContext) (bool, error) {
	out, err := expr.Run(p, contextEnv(ctx))
	if err != nil {
		return false, err
	}
	v, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("rule expression did not produce bool, got %T", out)
	}
	return v, nil
}

// envType returns a sentinel struct with the same shape used at runtime,
// purely for compile-time type checking.
func envType() map[string]any {
	return map[string]any{
		"user_id":                    int64(0),
		"user_tier":                  "",
		"user_register_at":           int64(0),
		"user_country":               "",
		"activity_id":                int64(0),
		"activity_type":              "",
		"current_time":               int64(0),
		"participation_count_today":  int64(0),
		"participation_count_total":  int64(0),
		"device_id":                  "",
		"ip":                         "",
		"custom_kv":                  map[string]string{},
	}
}

func contextEnv(c *rule.RuleContext) map[string]any {
	if c == nil {
		c = &rule.RuleContext{}
	}
	now := c.CurrentTime
	if now == 0 {
		now = time.Now().Unix()
	}
	customKv := c.CustomKv
	if customKv == nil {
		customKv = map[string]string{}
	}
	return map[string]any{
		"user_id":                   c.UserId,
		"user_tier":                 c.UserTier,
		"user_register_at":          c.UserRegisterAt,
		"user_country":              c.UserCountry,
		"activity_id":               c.ActivityId,
		"activity_type":             c.ActivityType,
		"current_time":              now,
		"participation_count_today": c.ParticipationCountToday,
		"participation_count_total": c.ParticipationCountTotal,
		"device_id":                 c.DeviceId,
		"ip":                        c.Ip,
		"custom_kv":                 customKv,
	}
}

// ===== custom helpers =====

func funcNow(...any) (any, error) {
	return time.Now().Unix(), nil
}

// duration("30d") / duration("24h") -> seconds
func funcDuration(args ...any) (any, error) {
	if len(args) != 1 {
		return int64(0), fmt.Errorf("duration takes 1 arg")
	}
	s, ok := args[0].(string)
	if !ok {
		return int64(0), fmt.Errorf("duration arg must be string")
	}
	if l := len(s); l > 1 && s[l-1] == 'd' {
		var n int64
		if _, err := fmt.Sscanf(s[:l-1], "%d", &n); err != nil {
			return int64(0), err
		}
		return n * 86400, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return int64(0), err
	}
	return int64(d.Seconds()), nil
}

func funcInSet(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("inSet takes 2 args")
	}
	v, _ := args[0].(string)
	list, _ := args[1].([]string)
	for _, e := range list {
		if e == v {
			return true, nil
		}
	}
	return false, nil
}
