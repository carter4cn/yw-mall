package logic

import (
	"context"
	"encoding/json"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type DestroySessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDestroySessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DestroySessionLogic {
	return &DestroySessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DestroySession revokes a single access token (logout). Idempotent: missing
// keys are treated as success so retries from a flaky client don't 500.
func (l *DestroySessionLogic) DestroySession(in *user.DestroySessionReq) (*user.Empty, error) {
	if in.AccessToken == "" {
		return &user.Empty{}, nil
	}

	raw, err := l.svcCtx.Redis.Get(l.ctx, sessionKey(in.AccessToken)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if err != redis.Nil {
		var sess sessionPayload
		if jerr := json.Unmarshal([]byte(raw), &sess); jerr == nil {
			if sess.RefreshToken != "" {
				_ = l.svcCtx.Redis.Del(l.ctx, refreshKey(sess.RefreshToken)).Err()
			}
			if sess.Uid > 0 {
				_ = l.svcCtx.Redis.SRem(l.ctx, userSessionsKey(sess.Uid), in.AccessToken).Err()
			}
		}
	}

	_ = l.svcCtx.Redis.Del(l.ctx, sessionKey(in.AccessToken)).Err()
	return &user.Empty{}, nil
}
