package logic

import (
	"context"
	"time"

	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/logx"
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

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	u, err := l.lookupUser(in.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": u.Id,
		"iat": now,
		"exp": now + 86400*7,
	}).SignedString([]byte(l.svcCtx.JwtSecretHot.Get()))
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{Id: int64(u.Id), Token: token}, nil
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
