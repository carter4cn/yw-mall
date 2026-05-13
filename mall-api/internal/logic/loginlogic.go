// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login is the legacy /api/user/login compat layer. After the P0 revamp the
// returned `token` is the opaque access_token (not a JWT). Clients that need
// refresh/csrf/expiry should migrate to /api/auth/login.
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	l.Logger.Infow("legacy /api/user/login hit; migrate to /api/auth/login",
		logx.Field("username", req.Username))
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{
		Id:    res.Id,
		Token: res.Token,
	}, nil
}
