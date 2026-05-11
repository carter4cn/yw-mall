package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdminsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAdminsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminsLogic {
	return &ListAdminsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAdminsLogic) ListAdmins(in *user.ListAdminsReq) (*user.ListAdminsResp, error) {
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM admin_user"); err != nil {
		return nil, err
	}

	var rows []*adminRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, username, password_hash, email, role, COALESCE(permissions,'') AS permissions, status, create_time, update_time FROM admin_user ORDER BY id DESC LIMIT ? OFFSET ?",
		pageSize, offset); err != nil {
		return nil, err
	}

	out := make([]*user.AdminInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAdminInfo(r))
	}
	return &user.ListAdminsResp{Admins: out, Total: total}, nil
}
