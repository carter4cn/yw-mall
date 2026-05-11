package logic

import (
	"context"
	"errors"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetAdminByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAdminByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminByIdLogic {
	return &GetAdminByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAdminByIdLogic) GetAdminById(in *user.GetAdminByIdReq) (*user.AdminInfo, error) {
	var a adminRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &a,
		"SELECT id, username, password_hash, email, role, COALESCE(permissions,'') AS permissions, status, create_time, update_time FROM admin_user WHERE id=? LIMIT 1",
		in.Id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return toAdminInfo(&a), nil
}
