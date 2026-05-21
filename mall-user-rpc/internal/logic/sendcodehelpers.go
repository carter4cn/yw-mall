// mall-user-rpc/internal/logic/sendcodehelpers.go
package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"mall-user-rpc/internal/svc"
)

const (
	verifyCodeTTL      = 5 * time.Minute
	verifyResendWindow = 60 * time.Second
)

type verifyPayload struct {
	Code           string `json:"code"`
	ChallengeToken string `json:"token"`
	SentAt         int64  `json:"sent_at"`
}

func verifyKey(scene int32, target string) string {
	return fmt.Sprintf("verify:%d:%s", scene, target)
}

func newChallengeTokenStr() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomDigit6() (string, error) {
	const digits = "0123456789"
	out := make([]byte, 6)
	max := big.NewInt(10)
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = digits[n.Int64()]
	}
	return string(out), nil
}

// storeVerifyCode 把 (code, token) 写 Redis，TTL 5min。
// 同 target 60 秒内再次发送会被拒绝。
func storeVerifyCode(ctx context.Context, svcCtx *svc.ServiceContext, scene int32, target, code, token string) error {
	k := verifyKey(scene, target)
	if raw, _ := svcCtx.Redis.Get(ctx, k).Bytes(); len(raw) > 0 {
		var old verifyPayload
		if json.Unmarshal(raw, &old) == nil &&
			time.Since(time.Unix(old.SentAt, 0)) < verifyResendWindow {
			return errors.New("发送太频繁，请稍后再试")
		}
	}
	p := verifyPayload{Code: code, ChallengeToken: token, SentAt: time.Now().Unix()}
	data, _ := json.Marshal(p)
	return svcCtx.Redis.Set(ctx, k, data, verifyCodeTTL).Err()
}

// consumeVerifyCode 校验 (token, code) 是否匹配 + 未过期，匹配则删除 (单次消费)。
func consumeVerifyCode(ctx context.Context, svcCtx *svc.ServiceContext, scene int32, target, code, token string) error {
	k := verifyKey(scene, target)
	raw, err := svcCtx.Redis.Get(ctx, k).Bytes()
	if err != nil {
		return errors.New("验证码已过期或不存在")
	}
	var p verifyPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return errors.New("验证码已过期或不存在")
	}
	if p.ChallengeToken != token {
		return errors.New("challenge token 不匹配")
	}
	if p.Code != code {
		return errors.New("验证码不正确")
	}
	_ = svcCtx.Redis.Del(ctx, k).Err()
	return nil
}
