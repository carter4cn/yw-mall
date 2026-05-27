package logic

import (
	"context"
	"errors"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableStaffLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableStaffLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableStaffLogic {
	return &DisableStaffLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DisableStaff 冻结员工（status=0）。owner 不能冻结。
// 注：被冻结员工的现存 session 不会立即失效；下次 MerchantLogin 时
// queryStaffByUserId(status=1) 取不到记录 → 登录被拒。
// 立即下线需另调 user-rpc.DestroyAllUserSessions（P1 改进）。
func (l *DisableStaffLogic) DisableStaff(in *shop.DisableStaffReq) (*shop.OkResp, error) {
	var role string
	if err := l.svcCtx.UserDB.QueryRowCtx(l.ctx, &role,
		"SELECT role FROM merchant_staff WHERE id=? AND shop_id=?",
		in.StaffId, in.ShopId); err != nil {
		return nil, errors.New("staff not found")
	}
	if role == RoleOwner {
		return nil, errors.New("cannot disable owner")
	}
	if _, err := l.svcCtx.UserDB.ExecCtx(l.ctx,
		"UPDATE merchant_staff SET status=0 WHERE id=? AND shop_id=?",
		in.StaffId, in.ShopId); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
