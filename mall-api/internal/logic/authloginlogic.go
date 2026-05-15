package logic

import (
	"context"
	"strings"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/user"

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
	username := strings.TrimSpace(req.Username)
	ip := IPFromCtx(l.ctx)

	if err := CheckLoginLock(l.ctx, l.svcCtx, "user", username, ip); err != nil {
		return nil, err
	}

	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		MarkLoginFail(l.ctx, l.svcCtx, "user", username, ip)
		return nil, err
	}
	ClearLoginFail(l.ctx, l.svcCtx, "user", username, ip)

	return &types.AuthLoginResp{
		Uid:             res.Id,
		Username:        req.Username,
		AccessToken:     res.Token,
		RefreshToken:    res.RefreshToken,
		ExpiresIn:       res.ExpiresIn,
		CsrfToken:       res.CsrfToken,
		PasswordExpired: res.PasswordExpired,
	}, nil
}
