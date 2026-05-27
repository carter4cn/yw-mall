package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopStaffLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopStaffLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopStaffLogic {
	return &ListShopStaffLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListShopStaff 列出本店全员工，owner 排首位。
func (l *ListShopStaffLogic) ListShopStaff(in *shop.ListShopStaffReq) (*shop.ListShopStaffResp, error) {
	var rows []struct {
		Id       uint64 `db:"id"`
		UserId   uint64 `db:"user_id"`
		Username string `db:"username"`
		Role     string `db:"role"`
		Status   int32  `db:"status"`
		JoinedAt int64  `db:"joined_at"`
	}
	err := l.svcCtx.UserDB.QueryRowsCtx(l.ctx, &rows, `
        SELECT s.id, s.user_id, COALESCE(u.username,'') AS username,
               s.role, s.status, s.joined_at
        FROM merchant_staff s
        LEFT JOIN `+"`user`"+` u ON u.id = s.user_id
        WHERE s.shop_id = ?
        ORDER BY (s.role='owner') DESC, s.joined_at`, in.ShopId)
	if err != nil {
		return nil, err
	}
	out := make([]*shop.StaffInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, &shop.StaffInfo{
			Id: int64(r.Id), ShopId: in.ShopId, UserId: int64(r.UserId),
			Username: r.Username, Role: r.Role, Status: r.Status, JoinedAt: r.JoinedAt,
		})
	}
	return &shop.ListShopStaffResp{Items: out}, nil
}
