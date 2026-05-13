package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogoutLogic {
	return &AuthLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthLogout revokes the access token used to authenticate this very request.
// The middleware put it in context — no need to reparse the header.
func (l *AuthLogoutLogic) AuthLogout() (*types.AuthLogoutResp, error) {
	token := middleware.AccessTokenFromCtx(l.ctx)
	if token != "" {
		if _, err := l.svcCtx.UserRpc.DestroySession(l.ctx, &user.DestroySessionReq{
			AccessToken: token,
		}); err != nil {
			l.Logger.Errorf("DestroySession failed: %v", err)
		}
	}
	return &types.AuthLogoutResp{Ok: true}, nil
}
