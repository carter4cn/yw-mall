package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthRefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthRefreshLogic {
	return &AuthRefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthRefresh rotates a refresh_token into a new access/refresh pair via
// user-rpc.RefreshSession. The old pair is revoked atomically on the rpc side.
func (l *AuthRefreshLogic) AuthRefresh(req *types.AuthRefreshReq) (*types.AuthRefreshResp, error) {
	sess, err := l.svcCtx.UserRpc.RefreshSession(l.ctx, &user.RefreshSessionReq{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	return &types.AuthRefreshResp{
		Uid:          sess.Uid,
		Username:     sess.Username,
		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		ExpiresIn:    sess.ExpiresIn,
		CsrfToken:    sess.CsrfToken,
	}, nil
}
