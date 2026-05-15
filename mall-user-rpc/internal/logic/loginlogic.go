package logic

import (
	"context"
	"errors"

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
//
// S4.5 hardening: refuse login when status=2 (account erased per data/erase).
// S4.3 hardening: surface password_expired so the gateway can force a rotation.
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	u, err := l.lookupUser(in.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, err
	}

	// Account-erased gate (S4.5). user table column `status` is owned by an
	// earlier migration; 0=disabled, 1=active, 2=erased. We rely on a raw
	// SELECT because the goctl-generated model ignores the column.
	var status int32
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &status,
		"SELECT status FROM `user` WHERE id=? LIMIT 1", u.Id); err == nil {
		if status == 2 {
			return nil, errors.New("账号已注销，无法登录")
		}
		if status == 0 {
			return nil, errors.New("账号已停用")
		}
	}

	// last_password_change for S4.3 expiry. Column added by S4.3 DDL; if the
	// query errors out (column not present yet) we treat it as 0 = never.
	var lastChange int64
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &lastChange,
		"SELECT COALESCE(last_password_change,0) FROM `user` WHERE id=? LIMIT 1", u.Id)

	sess, err := NewCreateSessionLogic(l.ctx, l.svcCtx).CreateSession(&user.CreateSessionReq{
		Uid:      int64(u.Id),
		Username: u.Username,
		Role:     "user",
	})
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Id:              sess.Uid,
		Token:           sess.AccessToken,
		RefreshToken:    sess.RefreshToken,
		ExpiresIn:       sess.ExpiresIn,
		CsrfToken:       sess.CsrfToken,
		PasswordExpired: passwordExpired(lastChange, l.svcCtx.PasswordPolicy.MaxAgeDays),
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
