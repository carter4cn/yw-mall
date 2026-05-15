package logic

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"mall-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// S4.2 c-side failed-login lockout. Mirrors the admin gateway implementation
// but in a separate Redis namespace so a wrong c-side password doesn't lock
// out an identically-named admin.

const (
	lockMaxAttempts = 5
	lockWindow      = 30 * time.Minute
)

func failKey(scope, key string) string { return "login_fail:" + scope + ":" + key }

func CheckLoginLock(ctx context.Context, svcCtx *svc.ServiceContext, scope, username, ip string) error {
	if svcCtx.Redis == nil {
		return nil
	}
	usr, err := svcCtx.Redis.Get(ctx, failKey(scope, username)).Int64()
	if err != nil && err.Error() != "redis: nil" {
		return nil
	}
	ipCount, err := svcCtx.Redis.Get(ctx, failKey(scope, ip)).Int64()
	if err != nil && err.Error() != "redis: nil" {
		return nil
	}
	if usr >= lockMaxAttempts || ipCount >= lockMaxAttempts {
		ttl := svcCtx.Redis.TTL(ctx, failKey(scope, username)).Val()
		if ttl <= 0 {
			ttl = svcCtx.Redis.TTL(ctx, failKey(scope, ip)).Val()
		}
		remaining := int64(ttl.Seconds())
		if remaining <= 0 {
			remaining = int64(lockWindow.Seconds())
		}
		return fmt.Errorf("账号已锁定，请 %d 秒后重试", remaining)
	}
	return nil
}

func MarkLoginFail(ctx context.Context, svcCtx *svc.ServiceContext, scope, username, ip string) {
	if svcCtx.Redis == nil {
		return
	}
	pipe := svcCtx.Redis.Pipeline()
	if username != "" {
		pipe.Incr(ctx, failKey(scope, username))
		pipe.Expire(ctx, failKey(scope, username), lockWindow)
	}
	if ip != "" {
		pipe.Incr(ctx, failKey(scope, ip))
		pipe.Expire(ctx, failKey(scope, ip), lockWindow)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		logx.WithContext(ctx).Errorf("MarkLoginFail: %v", err)
	}
}

func ClearLoginFail(ctx context.Context, svcCtx *svc.ServiceContext, scope, username, ip string) {
	if svcCtx.Redis == nil {
		return
	}
	keys := []string{}
	if username != "" {
		keys = append(keys, failKey(scope, username))
	}
	if ip != "" {
		keys = append(keys, failKey(scope, ip))
	}
	if len(keys) > 0 {
		_ = svcCtx.Redis.Del(ctx, keys...).Err()
	}
}

// ClientIP extracts the best-effort client IP, identical to the admin-side helper.
func ClientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		if i := strings.Index(h, ","); i > 0 {
			return strings.TrimSpace(h[:i])
		}
		return strings.TrimSpace(h)
	}
	if h := r.Header.Get("X-Real-IP"); h != "" {
		return strings.TrimSpace(h)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// requestIPKey lets the handler stash the client IP into ctx so the login
// logic can pick it up without taking *http.Request as a parameter.
type requestIPKey struct{}

func WithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, requestIPKey{}, ip)
}

func IPFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(requestIPKey{}).(string); ok {
		return v
	}
	return ""
}
