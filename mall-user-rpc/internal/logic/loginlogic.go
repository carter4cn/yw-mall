package logic

import (
	"context"

	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login validates the credentials and mints an opaque-token session.
//
// The legacy JWT path is gone — the response now carries access_token,
// refresh_token, expires_in and csrf_token so mall-api can pass them straight
// through to the browser in one RPC. `token` stays populated (= access_token)
// for backward compatibility with the old `/api/user/login` contract.
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	u, err := l.lookupUser(in.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, err
	}

	sess, err := NewCreateSessionLogic(l.ctx, l.svcCtx).CreateSession(&user.CreateSessionReq{
		Uid:      int64(u.Id),
		Username: u.Username,
		Role:     "user",
	})
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Id:           sess.Uid,
		Token:        sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		ExpiresIn:    sess.ExpiresIn,
		CsrfToken:    sess.CsrfToken,
	}, nil
}

// lookupUser tries the cached path first, then falls back to a master-pinned
// read (FOR UPDATE pins to the writer hostgroup in ProxySQL) so a Register
// quickly followed by a Login is not defeated by replication lag and the
// resulting negative-cache poisoning that would otherwise lock the user out
// for the cache TTL.
func (l *LoginLogic) lookupUser(username string) (*model.User, error) {
	u, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
	if err == nil {
		return u, nil
	}
	if err != model.ErrNotFound && err != sqlc.ErrNotFound {
		return nil, err
	}
	var fresh model.User
	q := "SELECT id, username, password, phone, avatar, create_time, update_time FROM `user` WHERE username = ? LIMIT 1 FOR UPDATE"
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &fresh, q, username); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &fresh, nil
}
