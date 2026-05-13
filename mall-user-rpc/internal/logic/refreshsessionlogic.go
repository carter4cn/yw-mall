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

// ErrRefreshExhausted is returned when a refresh token has been rotated more
// than MaxRotateCount times; callers must force the user to re-login.
var ErrRefreshExhausted = errors.New("refresh rotate limit reached")

// ErrRefreshNotFound is returned when the refresh token has no live entry
// (expired, revoked, or never existed).
var ErrRefreshNotFound = errors.New("refresh token not found")

// ErrRefreshDeviceMismatch is returned when the caller's device id doesn't
// match the device the refresh token was bound to. Defensive against token theft.
var ErrRefreshDeviceMismatch = errors.New("refresh device mismatch")

type RefreshSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshSessionLogic {
	return &RefreshSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshSession rotates both access and refresh tokens atomically: the old
// pair is deleted and a fresh pair is minted. RotateCount caps lifetime use.
func (l *RefreshSessionLogic) RefreshSession(in *user.RefreshSessionReq) (*user.SessionInfo, error) {
	if in.RefreshToken == "" {
		return nil, ErrRefreshNotFound
	}

	raw, err := l.svcCtx.Redis.Get(l.ctx, refreshKey(in.RefreshToken)).Result()
	if err == redis.Nil {
		return nil, ErrRefreshNotFound
	}
	if err != nil {
		return nil, err
	}

	var old refreshPayload
	if err := json.Unmarshal([]byte(raw), &old); err != nil {
		return nil, err
	}

	// Device binding (best-effort): only enforce when both sides have a value.
	if old.DeviceId != "" && in.DeviceId != "" && old.DeviceId != in.DeviceId {
		return nil, ErrRefreshDeviceMismatch
	}

	if old.RotateCount >= maxRotate(l.svcCtx.Config.Session.MaxRotateCount) {
		// Burn the refresh token so a stolen one can't be reused.
		_ = l.svcCtx.Redis.Del(l.ctx, refreshKey(in.RefreshToken)).Err()
		return nil, ErrRefreshExhausted
	}

	// Mint the new pair.
	newAccess := randomToken()
	newRefresh := randomToken()
	csrf := randomToken()
	now := time.Now().Unix()

	sess := sessionPayload{
		Uid:          old.Uid,
		Username:     old.Username,
		Role:         old.Role,
		ShopId:       old.ShopId,
		DeviceId:     old.DeviceId,
		IP:           old.IP,
		CsrfToken:    csrf,
		LoginTime:    old.LoginTime,
		LastActive:   now,
		RefreshToken: newRefresh,
		Perms:        old.Perms,
	}
	sessData, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	newRefPayload := refreshPayload{
		Uid:         old.Uid,
		Username:    old.Username,
		Role:        old.Role,
		ShopId:      old.ShopId,
		DeviceId:    old.DeviceId,
		IP:          old.IP,
		AccessToken: newAccess,
		RotateCount: old.RotateCount + 1,
		LoginTime:   old.LoginTime,
		Perms:       old.Perms,
	}
	refData, err := json.Marshal(newRefPayload)
	if err != nil {
		return nil, err
	}

	accessTTLDur := accessTTL(l.svcCtx.Config.Session.AccessTTLSeconds)
	refreshTTLDur := refreshTTL(l.svcCtx.Config.Session.RefreshTTLSeconds)

	if err := l.svcCtx.Redis.Set(l.ctx, sessionKey(newAccess), sessData, accessTTLDur).Err(); err != nil {
		return nil, err
	}
	if err := l.svcCtx.Redis.Set(l.ctx, refreshKey(newRefresh), refData, refreshTTLDur).Err(); err != nil {
		_ = l.svcCtx.Redis.Del(l.ctx, sessionKey(newAccess)).Err()
		return nil, err
	}

	// Best-effort revoke of the old pair and index update.
	if old.AccessToken != "" {
		_ = l.svcCtx.Redis.Del(l.ctx, sessionKey(old.AccessToken)).Err()
		_ = l.svcCtx.Redis.SRem(l.ctx, userSessionsKey(old.Uid), old.AccessToken).Err()
	}
	_ = l.svcCtx.Redis.Del(l.ctx, refreshKey(in.RefreshToken)).Err()
	_ = l.svcCtx.Redis.SAdd(l.ctx, userSessionsKey(old.Uid), newAccess).Err()

	return &user.SessionInfo{
		Uid:          old.Uid,
		Username:     old.Username,
		Role:         old.Role,
		ShopId:       old.ShopId,
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    int32(accessTTLDur / time.Second),
		CsrfToken:    csrf,
		LoginTime:    old.LoginTime,
		Perms:        old.Perms,
	}, nil
}
