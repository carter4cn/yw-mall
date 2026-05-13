package logic

import (
	"context"
	"encoding/json"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateSession mints a fresh opaque access/refresh pair and persists them to
// Redis along with the user_sessions:{uid} index. Returns full SessionInfo so
// the caller (LoginLogic, future SMS/OAuth flows) only needs one RPC roundtrip.
func (l *CreateSessionLogic) CreateSession(in *user.CreateSessionReq) (*user.SessionInfo, error) {
	role := in.Role
	if role == "" {
		role = "user"
	}

	access := randomToken()
	refresh := randomToken()
	csrf := randomToken()
	now := time.Now().Unix()

	sess := sessionPayload{
		Uid:          in.Uid,
		Username:     in.Username,
		Role:         role,
		ShopId:       in.ShopId,
		DeviceId:     in.DeviceId,
		IP:           in.Ip,
		CsrfToken:    csrf,
		LoginTime:    now,
		LastActive:   now,
		RefreshToken: refresh,
	}
	sessData, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	rp := refreshPayload{
		Uid:         in.Uid,
		Username:    in.Username,
		Role:        role,
		ShopId:      in.ShopId,
		DeviceId:    in.DeviceId,
		IP:          in.Ip,
		AccessToken: access,
		RotateCount: 0,
		LoginTime:   now,
	}
	refData, err := json.Marshal(rp)
	if err != nil {
		return nil, err
	}

	accessTTLDur := accessTTL(l.svcCtx.Config.Session.AccessTTLSeconds)
	refreshTTLDur := refreshTTL(l.svcCtx.Config.Session.RefreshTTLSeconds)

	if err := l.svcCtx.Redis.Set(l.ctx, sessionKey(access), sessData, accessTTLDur).Err(); err != nil {
		return nil, err
	}
	if err := l.svcCtx.Redis.Set(l.ctx, refreshKey(refresh), refData, refreshTTLDur).Err(); err != nil {
		// Best-effort cleanup of the stranded session key.
		_ = l.svcCtx.Redis.Del(l.ctx, sessionKey(access)).Err()
		return nil, err
	}
	if err := l.svcCtx.Redis.SAdd(l.ctx, userSessionsKey(in.Uid), access).Err(); err != nil {
		l.Logger.Errorf("CreateSession: SAdd user_sessions:%d failed: %v", in.Uid, err)
	}

	return &user.SessionInfo{
		Uid:          in.Uid,
		Username:     in.Username,
		Role:         role,
		ShopId:       in.ShopId,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int32(accessTTLDur / time.Second),
		CsrfToken:    csrf,
		LoginTime:    now,
	}, nil
}
