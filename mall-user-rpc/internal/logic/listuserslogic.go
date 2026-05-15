package logic

import (
	"context"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type adminUserRow struct {
	Id         uint64    `db:"id"`
	Username   string    `db:"username"`
	Phone      string    `db:"phone"`
	Avatar     string    `db:"avatar"`
	CreateTime time.Time `db:"create_time"`
}

func (l *ListUsersLogic) ListUsers(in *user.ListUsersReq) (*user.ListUsersResp, error) {
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var (
		rows  []*adminUserRow
		total int64
	)
	if in.Keyword != "" {
		kw := "%" + in.Keyword + "%"
		if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM `user` WHERE username LIKE ? OR phone LIKE ?", kw, kw); err != nil {
			return nil, err
		}
		if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
			"SELECT id, username, phone, avatar, create_time FROM `user` WHERE username LIKE ? OR phone LIKE ? ORDER BY id DESC LIMIT ? OFFSET ?",
			kw, kw, pageSize, offset); err != nil {
			return nil, err
		}
	} else {
		if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM `user`"); err != nil {
			return nil, err
		}
		if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
			"SELECT id, username, phone, avatar, create_time FROM `user` ORDER BY id DESC LIMIT ? OFFSET ?",
			pageSize, offset); err != nil {
			return nil, err
		}
	}

	out := make([]*user.GetUserResp, 0, len(rows))
	for _, r := range rows {
		// S4.6 dual-read: decrypt v1: blobs, pass legacy plaintext through.
		// On decrypt error log + show empty so admin UI doesn't leak garbage.
		phone, err := cryptox.DecryptIfCiphertext(r.Phone)
		if err != nil {
			l.Logger.Errorf("ListUsers: decrypt phone for uid=%d failed: %v", r.Id, err)
			phone = ""
		}
		out = append(out, &user.GetUserResp{
			Id:         int64(r.Id),
			Username:   r.Username,
			Phone:      phone,
			Avatar:     r.Avatar,
			CreateTime: r.CreateTime.Unix(),
		})
	}
	return &user.ListUsersResp{Users: out, Total: total}, nil
}
