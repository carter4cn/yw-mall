package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VerifyToken validates HMAC and (optionally) consumes the token via Redis
// SETNX so a leaked token can be replayed at most once before being burned.
func (l *VerifyTokenLogic) VerifyToken(in *risk.VerifyTokenReq) (*risk.VerifyTokenResp, error) {
	parts := strings.SplitN(in.Token, ".", 2)
	if len(parts) != 2 {
		return &risk.VerifyTokenResp{Valid: false, Reason: "malformed"}, nil
	}
	bodyB64, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(l.svcCtx.Config.TokenSecret))
	mac.Write([]byte(bodyB64))
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(want)) {
		return &risk.VerifyTokenResp{Valid: false, Reason: "bad signature"}, nil
	}

	body, err := base64.RawURLEncoding.DecodeString(bodyB64)
	if err != nil {
		return &risk.VerifyTokenResp{Valid: false, Reason: "bad payload"}, nil
	}
	var p tokenPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return &risk.VerifyTokenResp{Valid: false, Reason: "bad payload"}, nil
	}
	if p.UserId != in.UserId || p.ActivityId != in.ActivityId {
		return &risk.VerifyTokenResp{Valid: false, Reason: "subject mismatch"}, nil
	}
	if time.Now().Unix() > p.ExpiresAt {
		return &risk.VerifyTokenResp{Valid: false, Reason: "expired"}, nil
	}
	if in.Consume {
		// SETNX with TTL covering the token's max remaining life — burns the jti.
		ttl := p.ExpiresAt - time.Now().Unix()
		if ttl < 1 {
			ttl = 1
		}
		ok, err := l.svcCtx.Redis.SetnxExCtx(l.ctx, "risk:tok:used:"+p.Jti, "1", int(ttl))
		if err != nil {
			return &risk.VerifyTokenResp{Valid: false, Reason: "redis: " + err.Error()}, nil
		}
		if !ok {
			return &risk.VerifyTokenResp{Valid: false, Reason: "already used"}, nil
		}
	}
	return &risk.VerifyTokenResp{Valid: true}, nil
}
