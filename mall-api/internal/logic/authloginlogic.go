package logic

import (
	"context"

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
func (l *AuthLoginLogic) AuthLogin(req *types.AuthLoginReq) (*types.AuthLoginResp, error) {
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.AuthLoginResp{
		Uid:          res.Id,
		Username:     req.Username,
		AccessToken:  res.Token,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
		CsrfToken:    res.CsrfToken,
	}, nil
}
