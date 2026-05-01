package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"mall-risk-rpc/internal/lua"
	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type RateLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRateLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RateLimitLogic {
	return &RateLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RateLimit runs the sliding-window Lua and returns allowed/remaining/reset.
// Caller-supplied window/max take precedence over rate_limit_config rows so
// callers can tighten on a hot activity without a config push.
func (l *RateLimitLogic) RateLimit(in *risk.RateLimitReq) (*risk.RateLimitResp, error) {
	maxn := in.MaxCount
	window := in.WindowSeconds
	if maxn <= 0 || window <= 0 {
		// fall back to default; future: load from rate_limit_config keyed by (activity_id, subject_type)
		maxn = 30
		window = 60
	}
	key := fmt.Sprintf("risk:rl:%d:%s:%s", in.ActivityId, in.SubjectType, in.SubjectValue)
	nowMs := time.Now().UnixMilli()
	v, err := l.svcCtx.Redis.EvalCtx(l.ctx, lua.SlidingWindow,
		[]string{key},
		strconv.FormatInt(nowMs, 10),
		strconv.Itoa(int(window)*1000),
		strconv.Itoa(int(maxn)),
	)
	if err != nil {
		return nil, err
	}
	arr, _ := v.([]any)
	if len(arr) < 3 {
		return &risk.RateLimitResp{Allowed: false, Remaining: 0, ResetAt: 0}, nil
	}
	allowed, _ := arr[0].(int64)
	remaining, _ := arr[1].(int64)
	resetAtMs, _ := arr[2].(int64)
	return &risk.RateLimitResp{
		Allowed:   allowed == 1,
		Remaining: int32(remaining),
		ResetAt:   resetAtMs / 1000,
	}, nil
}
