package logic

import (
	"context"
	"errors"
	"strings"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLoginLogic {
	return &AuthLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthLogin delegates credential check + session minting to user-rpc.Login,
// which (after the P0 revamp) returns the full session in one roundtrip.
//
// S4.2 / S4.3 hardening:
//   - failed-login lock counters in Redis (5 in 30 min)
//   - surfaces password_expired so FE can force a rotation
func (l *AuthLoginLogic) AuthLogin(req *types.AuthLoginReq) (*types.AuthLoginResp, error) {
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = strings.TrimSpace(req.Username) // 兼容老 FE
	}
	if account == "" {
		return nil, errors.New("account required")
	}
	ip := IPFromCtx(l.ctx)

	if err := CheckLoginLock(l.ctx, l.svcCtx, "user", account, ip); err != nil {
		return nil, err
	}

	res, err := l.svcCtx.UserRpc.LoginV2(l.ctx, &userclient.LoginV2Req{
		Account:  account,
		Password: req.Password,
	})
	if err != nil {
		MarkLoginFail(l.ctx, l.svcCtx, "user", account, ip)
		return nil, err
	}
	ClearLoginFail(l.ctx, l.svcCtx, "user", account, ip)

	return &types.AuthLoginResp{
		Uid:             res.Id,
		Username:        account,
		AccessToken:     res.Token,
		RefreshToken:    res.RefreshToken,
		ExpiresIn:       res.ExpiresIn,
		CsrfToken:       res.CsrfToken,
		PasswordExpired: res.PasswordExpired,
	}, nil
}
