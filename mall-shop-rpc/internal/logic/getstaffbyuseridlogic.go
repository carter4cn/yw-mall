package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStaffByUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStaffByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStaffByUserIdLogic {
	return &GetStaffByUserIdLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetStaffByUserId 给 admin-api 的 MerchantLogin 用：拿 (shopId, role, perms)
// 三元组以写入 session.Perms。Found=false 表示该用户没绑任何 shop。
func (l *GetStaffByUserIdLogic) GetStaffByUserId(in *shop.GetStaffByUserIdReq) (*shop.GetStaffByUserIdResp, error) {
	s, err := queryStaffByUserId(l.ctx, l.svcCtx, in.UserId)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return &shop.GetStaffByUserIdResp{Found: false}, nil
	}
	var username string
	_ = l.svcCtx.UserDB.QueryRowCtx(l.ctx, &username,
		"SELECT username FROM `user` WHERE id=?", s.UserId)
	return &shop.GetStaffByUserIdResp{
		Found: true,
		Staff: &shop.StaffInfo{
			Id: int64(s.Id), ShopId: int64(s.ShopId), UserId: int64(s.UserId),
			Username: username, Role: s.Role, Status: s.Status, JoinedAt: s.JoinedAt,
		},
		Perms: RolePerms[s.Role],
	}, nil
}
