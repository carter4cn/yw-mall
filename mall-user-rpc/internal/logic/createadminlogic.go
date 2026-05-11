package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type CreateAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAdminLogic {
	return &CreateAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAdminLogic) CreateAdmin(in *user.CreateAdminReq) (*user.CreateAdminResp, error) {
	if strings.TrimSpace(in.Username) == "" || strings.TrimSpace(in.Password) == "" {
		return nil, errors.New("username and password are required")
	}
	role := in.Role
	if role == "" {
		role = "admin"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO admin_user (username, password_hash, email, role, permissions, status, create_time, update_time) VALUES (?, ?, ?, ?, ?, 1, ?, ?)",
		in.Username, string(hash), in.Email, role, in.Permissions, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &user.CreateAdminResp{Id: id}, nil
}
