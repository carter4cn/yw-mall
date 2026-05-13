package middleware

import (
	"context"
	"net/http"
	"strings"

	"mall-user-rpc/userclient"
	userpb "mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

// Context-key types: unexported so other packages can only set/read uid through
// the helpers in this file. Avoids the classic string-collision bug.
type (
	sessionUidKey      struct{}
	sessionTokenKey    struct{}
	sessionUsernameKey struct{}
	sessionRoleKey     struct{}
	sessionCsrfKey     struct{}
)

// WithSession injects a validated session's uid + token (+ a few useful side
// values) into the request context so downstream logic can read them via the
// helpers below.
func WithSession(ctx context.Context, sess *userpb.SessionInfo, accessToken string) context.Context {
	ctx = context.WithValue(ctx, sessionUidKey{}, sess.Uid)
	ctx = context.WithValue(ctx, sessionTokenKey{}, accessToken)
	ctx = context.WithValue(ctx, sessionUsernameKey{}, sess.Username)
	ctx = context.WithValue(ctx, sessionRoleKey{}, sess.Role)
	ctx = context.WithValue(ctx, sessionCsrfKey{}, sess.CsrfToken)
	return ctx
}

// UidFromCtx returns the authenticated uid, or 0 if absent. Logic-layer callers
// MUST treat 0 as unauthenticated even though the middleware should have
// already rejected anonymous requests — defence in depth.
func UidFromCtx(ctx context.Context) int64 {
	if v, ok := ctx.Value(sessionUidKey{}).(int64); ok {
		return v
	}
	return 0
}

// AccessTokenFromCtx returns the raw bearer token used for the current request.
// Used by logout to revoke without re-reading the header.
func AccessTokenFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(sessionTokenKey{}).(string); ok {
		return v
	}
	return ""
}

// CsrfTokenFromCtx returns the session-bound CSRF token, used by future
// double-submit checks on writes.
func CsrfTokenFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(sessionCsrfKey{}).(string); ok {
		return v
	}
	return ""
}

// NewSessionAuthMiddleware wires a Redis-backed opaque-token check that
// replaces go-zero's `rest.WithJwt`. We take the UserRpc client directly
// instead of *svc.ServiceContext to avoid an import cycle (svc imports
// middleware to expose `AdminToken`).
func NewSessionAuthMiddleware(userRpc userclient.User) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			sess, err := userRpc.ValidateSession(r.Context(), &userpb.ValidateSessionReq{
				AccessToken: token,
			})
			if err != nil || sess == nil || sess.Uid <= 0 {
				if err != nil {
					logx.WithContext(r.Context()).Errorf("ValidateSession failed: %v", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next(w, r.WithContext(WithSession(r.Context(), sess, token)))
		}
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	// Tolerate both `Bearer xxx` and `bearer xxx`.
	if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}
