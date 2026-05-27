package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeInvitationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeInvitationLogic {
	return &RevokeInvitationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// RevokeInvitation 软删 status=3，仅对 status=0（pending）的条目生效。
// 已 accepted/expired 的不动。
func (l *RevokeInvitationLogic) RevokeInvitation(in *shop.RevokeInvitationReq) (*shop.OkResp, error) {
	if _, err := l.svcCtx.UserDB.ExecCtx(l.ctx,
		"UPDATE merchant_staff_invitation SET status=3 WHERE id=? AND shop_id=? AND status=0",
		in.InvitationId, in.ShopId); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
