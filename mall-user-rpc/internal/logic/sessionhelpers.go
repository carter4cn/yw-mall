package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"mall-user-rpc/internal/svc"

	"github.com/redis/go-redis/v9"
)

// Session storage layout (Redis):
//
//   session:{access_token}        JSON{uid,username,role,shopId,deviceId,ip,
//                                       csrfToken,loginTime,lastActive,
//                                       refreshToken}             TTL access
//   refresh:{refresh_token}       JSON{uid,deviceId,accessToken,rotateCount}
//                                                                 TTL refresh
//   user_sessions:{uid}           SET<access_token>               no TTL
//
// All TTLs come from svc.Config.Session (with safe defaults below).

const (
	defaultAccessTTLSeconds  = int64(30 * 60)        // 30 min
	defaultRefreshTTLSeconds = int64(7 * 24 * 3600)  // 7 days
	defaultMaxRotateCount    = int32(10)
)

// sessionPayload is what we store under session:{access_token}.
type sessionPayload struct {
	Uid          int64    `json:"uid"`
	Username     string   `json:"username"`
	Role         string   `json:"role"`
	ShopId       int64    `json:"shopId,omitempty"`
	DeviceId     string   `json:"deviceId,omitempty"`
	IP           string   `json:"ip,omitempty"`
	CsrfToken    string   `json:"csrfToken"`
	LoginTime    int64    `json:"loginTime"`
	LastActive   int64    `json:"lastActive"`
	RefreshToken string   `json:"refreshToken"`
	Perms        []string `json:"perms,omitempty"`
}

// refreshPayload is what we store under refresh:{refresh_token}.
type refreshPayload struct {
	Uid         int64    `json:"uid"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	ShopId      int64    `json:"shopId,omitempty"`
	DeviceId    string   `json:"deviceId,omitempty"`
	IP          string   `json:"ip,omitempty"`
	AccessToken string   `json:"accessToken"`
	RotateCount int32    `json:"rotateCount"`
	LoginTime   int64    `json:"loginTime"`
	Perms       []string `json:"perms,omitempty"`
}

// randomToken returns 32 random bytes encoded as URL-safe base64 (no padding).
// 256 bits of entropy is well above the 128-bit floor the design doc calls for.
func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func sessionKey(accessToken string) string {
	return "session:" + accessToken
}

func refreshKey(refreshToken string) string {
	return "refresh:" + refreshToken
}

func userSessionsKey(uid int64) string {
	return fmt.Sprintf("user_sessions:%d", uid)
}

// accessTTL returns the configured access-token TTL with a safe default.
func accessTTL(seconds int64) time.Duration {
	if seconds <= 0 {
		seconds = defaultAccessTTLSeconds
	}
	return time.Duration(seconds) * time.Second
}

// refreshTTL returns the configured refresh-token TTL with a safe default.
func refreshTTL(seconds int64) time.Duration {
	if seconds <= 0 {
		seconds = defaultRefreshTTLSeconds
	}
	return time.Duration(seconds) * time.Second
}

func maxRotate(n int32) int32 {
	if n <= 0 {
		return defaultMaxRotateCount
	}
	return n
}

// destroyUserSessionsByRole wipes every session for {uid} whose payload Role
// matches `role`. Used after password change/reset so a credential rotation
// kicks the matching identity off all devices without nuking the other identity
// that may share the same uid (admin_user.id and user.id collide in this
// codebase — both write to user_sessions:{uid}).
//
// Best-effort: returns the count it removed plus the first error encountered;
// callers typically only log it.
func destroyUserSessionsByRole(ctx context.Context, svcCtx *svc.ServiceContext, uid int64, role string) (int, error) {
	if uid <= 0 || role == "" {
		return 0, nil
	}
	tokens, err := svcCtx.Redis.SMembers(ctx, userSessionsKey(uid)).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	removed := 0
	for _, access := range tokens {
		raw, gerr := svcCtx.Redis.Get(ctx, sessionKey(access)).Result()
		if gerr != nil {
			// session already expired/missing — drop the dangling index entry too
			_ = svcCtx.Redis.SRem(ctx, userSessionsKey(uid), access).Err()
			continue
		}
		var sess sessionPayload
		if jerr := json.Unmarshal([]byte(raw), &sess); jerr != nil {
			continue
		}
		if sess.Role != role {
			continue
		}
		if sess.RefreshToken != "" {
			_ = svcCtx.Redis.Del(ctx, refreshKey(sess.RefreshToken)).Err()
		}
		_ = svcCtx.Redis.Del(ctx, sessionKey(access)).Err()
		_ = svcCtx.Redis.SRem(ctx, userSessionsKey(uid), access).Err()
		removed++
	}
	return removed, nil
}
