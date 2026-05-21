// mall-user-rpc/internal/logic/sendverifycodelogic.go
package logic

import (
	"context"
	"errors"
	"fmt"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerifyCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendVerifyCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerifyCodeLogic {
	return &SendVerifyCodeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

const (
	channelSms   = 1
	channelEmail = 2
)

func (l *SendVerifyCodeLogic) SendVerifyCode(in *user.SendVerifyCodeReq) (*user.SendVerifyCodeResp, error) {
	if in.Target == "" {
		return nil, errors.New("target required")
	}
	code, err := randomDigit6()
	if err != nil {
		return nil, err
	}
	token, err := newChallengeTokenStr()
	if err != nil {
		return nil, err
	}
	if err := storeVerifyCode(l.ctx, l.svcCtx, in.Scene, in.Target, code, token); err != nil {
		return nil, err
	}

	switch in.Channel {
	case channelSms:
		// 复用 S4 现成 SmsSend 实现：写入相同 Redis key 那套是 admin MFA 专用,
		// 这里直接 logx 占位，跟 EmailSend 同模式，避免双写 Redis
		logx.WithContext(l.ctx).Infof("[mock-sms] scene=%d target=%s code=%s", in.Scene, in.Target, code)
	case channelEmail:
		if err := EmailSend(l.ctx, in.Target, code); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported channel: %d", in.Channel)
	}

	return &user.SendVerifyCodeResp{
		ChallengeToken: token,
		ExpiresIn:      int32(verifyCodeTTL.Seconds()),
		DevCode:        code, // mock 阶段回显，接真 SMS 时改返空串
	}, nil
}
