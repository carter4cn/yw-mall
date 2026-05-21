// mall-user-rpc/internal/logic/loginv2logic.go
package logic

import (
	"context"
	"errors"
	"strings"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type LoginV2Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginV2Logic {
	return &LoginV2Logic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *LoginV2Logic) LoginV2(in *user.LoginV2Req) (*user.LoginResp, error) {
	account := strings.TrimSpace(in.Account)
	password := trimPlain(in.Password)
	if account == "" || password == "" {
		return nil, errors.New("用户名或密码错误")
	}

	var row struct {
		Id                 uint64 `db:"id"`
		Username           string `db:"username"`
		PasswordHash       string `db:"password"`
		LastPasswordChange int64  `db:"last_password_change"`
		Status             int32  `db:"status"`
	}

	var err error
	switch {
	case emailRE.MatchString(account):
		hash := cryptox.Hmac(strings.ToLower(account))
		err = l.svcCtx.DB.QueryRowCtx(l.ctx, &row, `
            SELECT id, username, password, last_password_change, status
            FROM user WHERE email_hash=? LIMIT 1`, hash)
	case phoneRE.MatchString(account):
		hash := cryptox.Hmac(account)
		err = l.svcCtx.DB.QueryRowCtx(l.ctx, &row, `
            SELECT id, username, password, last_password_change, status
            FROM user WHERE phone_hash=? LIMIT 1`, hash)
	default:
		err = l.svcCtx.DB.QueryRowCtx(l.ctx, &row, `
            SELECT id, username, password, last_password_change, status
            FROM user WHERE username=? LIMIT 1`, account)
	}

	if err != nil {
		if err == sqlx.ErrNotFound {
			// 防枚举：用户不存在 → 同一错误信息
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	if row.Status == 0 {
		return nil, errors.New("账号已停用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	expired := passwordExpired(row.LastPasswordChange, l.svcCtx.PasswordPolicy.MaxAgeDays)

	// 复用 P0 CreateSession
	sess, err := NewCreateSessionLogic(l.ctx, l.svcCtx).CreateSession(&user.CreateSessionReq{
		Uid:      int64(row.Id),
		Username: row.Username,
		Role:     "user",
	})
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Id:              int64(row.Id),
		Token:           sess.AccessToken,
		RefreshToken:    sess.RefreshToken,
		ExpiresIn:       sess.ExpiresIn,
		CsrfToken:       sess.CsrfToken,
		PasswordExpired: expired,
	}, nil
}
