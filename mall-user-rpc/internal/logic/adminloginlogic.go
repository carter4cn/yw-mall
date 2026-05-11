package logic

import (
	"context"
	"errors"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AdminLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminLoginLogic) AdminLogin(in *user.AdminLoginReq) (*user.AdminLoginResp, error) {
	var a adminRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &a,
		"SELECT id, username, password_hash, email, role, COALESCE(permissions,'') AS permissions, status, create_time, update_time FROM admin_user WHERE username=? LIMIT 1",
		in.Username)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	if a.Status != 1 {
		return nil, errors.New("admin disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errors.New("wrong password")
	}
	return &user.AdminLoginResp{
		Id:          int64(a.Id),
		Username:    a.Username,
		Role:        a.Role,
		Permissions: a.Permissions,
	}, nil
}
