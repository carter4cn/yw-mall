package logic

import (
	"context"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPendingInvitationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPendingInvitationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPendingInvitationsLogic {
	return &ListPendingInvitationsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListPendingInvitations 仅返 status=0 且 expires_at > now 的邀请。
// 旧的过期邀请由 ListPendingInvitations 隐式过滤；
// 真正 GC 是 status=2/3 标记 + 后续清表脚本（本期不做）。
func (l *ListPendingInvitationsLogic) ListPendingInvitations(in *shop.ListPendingInvitationsReq) (*shop.ListPendingInvitationsResp, error) {
	var rows []struct {
		Id          uint64 `db:"id"`
		TargetPhone string `db:"target_phone"`
		TargetEmail string `db:"target_email"`
		Role        string `db:"role"`
		Status      int32  `db:"status"`
		ExpiresAt   int64  `db:"expires_at"`
		CreateTime  int64  `db:"create_time_unix"`
	}
	err := l.svcCtx.UserDB.QueryRowsCtx(l.ctx, &rows, `
        SELECT id, target_phone, target_email, role, status, expires_at,
               UNIX_TIMESTAMP(create_time) AS create_time_unix
        FROM merchant_staff_invitation
        WHERE shop_id=? AND status=0 AND expires_at>?
        ORDER BY id DESC`, in.ShopId, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	out := make([]*shop.InvitationInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, &shop.InvitationInfo{
			Id: int64(r.Id), ShopId: in.ShopId,
			TargetPhone: r.TargetPhone, TargetEmail: r.TargetEmail,
			Role: r.Role, Status: r.Status,
			ExpiresAt: r.ExpiresAt, CreateTime: r.CreateTime,
		})
	}
	return &shop.ListPendingInvitationsResp{Items: out}, nil
}
