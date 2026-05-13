package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// ErrSessionNotFound is returned when an access token has no live Redis entry
// (expired, logged out, or never existed). Callers (mall-api middleware) map
// this to HTTP 401.
var ErrSessionNotFound = errors.New("session not found")

type ValidateSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateSessionLogic {
	return &ValidateSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateSession reads session:{access_token}, refreshes the access TTL
// (sliding expiry), and returns the embedded SessionInfo. Missing keys return
// ErrSessionNotFound; the mall-api middleware converts that into 401.
func (l *ValidateSessionLogic) ValidateSession(in *user.ValidateSessionReq) (*user.SessionInfo, error) {
	if in.AccessToken == "" {
		return nil, ErrSessionNotFound
	}
	key := sessionKey(in.AccessToken)
	raw, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	var sess sessionPayload
	if err := json.Unmarshal([]byte(raw), &sess); err != nil {
		return nil, err
	}

	// Sliding renewal — push the TTL back to the full window on every access.
	ttl := accessTTL(l.svcCtx.Config.Session.AccessTTLSeconds)
	now := time.Now().Unix()
	sess.LastActive = now
	if updated, err := json.Marshal(sess); err == nil {
		if err := l.svcCtx.Redis.Set(l.ctx, key, updated, ttl).Err(); err != nil {
			l.Logger.Errorf("ValidateSession: refresh TTL failed: %v", err)
		}
	} else {
		// Fall back to plain EXPIRE if marshalling the activity timestamp fails.
		_ = l.svcCtx.Redis.Expire(l.ctx, key, ttl).Err()
	}

	return &user.SessionInfo{
		Uid:          sess.Uid,
		Username:     sess.Username,
		Role:         sess.Role,
		ShopId:       sess.ShopId,
		AccessToken:  in.AccessToken,
		RefreshToken: sess.RefreshToken,
		ExpiresIn:    int32(ttl / time.Second),
		CsrfToken:    sess.CsrfToken,
		LoginTime:    sess.LoginTime,
		Perms:        sess.Perms,
	}, nil
}
