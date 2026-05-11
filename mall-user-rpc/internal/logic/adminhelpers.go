package logic

import (
	"mall-user-rpc/user"
)

type adminRow struct {
	Id           uint64 `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Email        string `db:"email"`
	Role         string `db:"role"`
	Permissions  string `db:"permissions"`
	Status       int64  `db:"status"`
	CreateTime   int64  `db:"create_time"`
	UpdateTime   int64  `db:"update_time"`
}

func toAdminInfo(a *adminRow) *user.AdminInfo {
	return &user.AdminInfo{
		Id:          int64(a.Id),
		Username:    a.Username,
		Email:       a.Email,
		Role:        a.Role,
		Permissions: a.Permissions,
		Status:      int32(a.Status),
		CreateTime:  a.CreateTime,
	}
}
