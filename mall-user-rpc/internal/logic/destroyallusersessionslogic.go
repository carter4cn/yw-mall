package logic

import (
	"context"
	"encoding/json"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type DestroyAllUserSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDestroyAllUserSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DestroyAllUserSessionsLogic {
	return &DestroyAllUserSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DestroyAllUserSessions wipes every session bound to a user — used for
// password reset, account ban, and the "kick all my devices" feature.
func (l *DestroyAllUserSessionsLogic) DestroyAllUserSessions(in *user.DestroyAllUserSessionsReq) (*user.Empty, error) {
	if in.Uid <= 0 {
		return &user.Empty{}, nil
	}

	tokens, err := l.svcCtx.Redis.SMembers(l.ctx, userSessionsKey(in.Uid)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for _, access := range tokens {
		// Look up the embedded refresh token so we can revoke it alongside.
		if raw, gerr := l.svcCtx.Redis.Get(l.ctx, sessionKey(access)).Result(); gerr == nil {
			var sess sessionPayload
			if jerr := json.Unmarshal([]byte(raw), &sess); jerr == nil && sess.RefreshToken != "" {
				_ = l.svcCtx.Redis.Del(l.ctx, refreshKey(sess.RefreshToken)).Err()
			}
		}
		_ = l.svcCtx.Redis.Del(l.ctx, sessionKey(access)).Err()
	}
	_ = l.svcCtx.Redis.Del(l.ctx, userSessionsKey(in.Uid)).Err()
	return &user.Empty{}, nil
}
