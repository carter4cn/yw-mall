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

	// S4.3 expiry hint. Column added by S4.3 DDL; if absent, default 0 = never.
	var lastChange int64
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &lastChange,
		"SELECT COALESCE(last_password_change,0) FROM admin_user WHERE id=? LIMIT 1", a.Id)

	// S4.1 MFA hint. Gateway uses this to issue the challenge token instead of
	// minting a session right away. Column belongs to admin_mfa.enabled.
	var mfaEnabled int32
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &mfaEnabled,
		"SELECT enabled FROM admin_mfa WHERE admin_id=? LIMIT 1", a.Id)

	return &user.AdminLoginResp{
		Id:              int64(a.Id),
		Username:        a.Username,
		Role:            a.Role,
		Permissions:     a.Permissions,
		PasswordExpired: passwordExpired(lastChange, l.svcCtx.PasswordPolicy.MaxAgeDays),
		MfaRequired:     mfaEnabled == 1,
	}, nil
}
