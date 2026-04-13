package logic

import (
	"context"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/golang-jwt/jwt/v4"
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
	u, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password))
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": u.Id,
		"iat": now,
		"exp": now + 86400*7,
	}).SignedString([]byte(l.svcCtx.Config.JwtAuth.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{Id: int64(u.Id), Token: token}, nil
}
