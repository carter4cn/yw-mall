package logic

import (
	"context"
	"errors"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateStaffRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStaffRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStaffRoleLogic {
	return &UpdateStaffRoleLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateStaffRole 改员工角色。owner 不能改，也不能转让 owner。
func (l *UpdateStaffRoleLogic) UpdateStaffRole(in *shop.UpdateStaffRoleReq) (*shop.OkResp, error) {
	if !IsValidRole(in.NewRole) {
		return nil, errors.New("invalid role")
	}
	if in.NewRole == RoleOwner {
		return nil, errors.New("cannot transfer owner via this endpoint")
	}
	var currentRole string
	if err := l.svcCtx.UserDB.QueryRowCtx(l.ctx, &currentRole,
		"SELECT role FROM merchant_staff WHERE id=? AND shop_id=?",
		in.StaffId, in.ShopId); err != nil {
		return nil, errors.New("staff not found")
	}
	if currentRole == RoleOwner {
		return nil, errors.New("cannot change owner role here")
	}
	if _, err := l.svcCtx.UserDB.ExecCtx(l.ctx,
		"UPDATE merchant_staff SET role=? WHERE id=? AND shop_id=?",
		in.NewRole, in.StaffId, in.ShopId); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
