package logic

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type IssueTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIssueTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IssueTokenLogic {
	return &IssueTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// tokenPayload is the cleartext claims block. The on-wire token is
// base64(payload).hex(hmac) — compact and trivially parseable in Go.
type tokenPayload struct {
	Jti        string `json:"jti"`
	UserId     int64  `json:"uid"`
	ActivityId int64  `json:"aid"`
	ExpiresAt  int64  `json:"exp"`
}

func (l *IssueTokenLogic) IssueToken(in *risk.IssueTokenReq) (*risk.IssueTokenResp, error) {
	ttl := in.TtlSeconds
	if ttl <= 0 {
		ttl = 300
	}
	jtiBytes := make([]byte, 12)
	if _, err := rand.Read(jtiBytes); err != nil {
		return nil, err
	}
	jti := hex.EncodeToString(jtiBytes)
	exp := time.Now().Unix() + int64(ttl)
	payload := tokenPayload{Jti: jti, UserId: in.UserId, ActivityId: in.ActivityId, ExpiresAt: exp}
	body, _ := json.Marshal(payload)
	bodyB64 := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(l.svcCtx.Config.TokenSecret))
	mac.Write([]byte(bodyB64))
	sig := hex.EncodeToString(mac.Sum(nil))
	token := bodyB64 + "." + sig

	return &risk.IssueTokenResp{Token: token, ExpiresAt: exp}, nil
}

// SignAndPack is exported for testing/round-tripping. Not used by the RPC.
var _ = fmt.Sprintf
