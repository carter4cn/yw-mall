package logic

import (
	"context"
	"errors"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AcceptInvitationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptInvitationLogic {
	return &AcceptInvitationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AcceptInvitation 校验 code 未过期 + 未消费 + acceptor phone/email
// 与 target 匹配（防代领），然后 INSERT staff + UPDATE invitation.
// 已绑此店报错。
func (l *AcceptInvitationLogic) AcceptInvitation(in *shop.AcceptInvitationReq) (*shop.AcceptInvitationResp, error) {
	if in.InvitationCode == "" || in.AcceptorUid <= 0 {
		return nil, errors.New("invalid request")
	}
	var inv struct {
		Id          uint64 `db:"id"`
		ShopId      uint64 `db:"shop_id"`
		TargetPhone string `db:"target_phone"`
		TargetEmail string `db:"target_email"`
		Role        string `db:"role"`
		ExpiresAt   int64  `db:"expires_at"`
		InvitedBy   uint64 `db:"invited_by"`
	}
	err := l.svcCtx.UserDB.QueryRowCtx(l.ctx, &inv, `
        SELECT id, shop_id, target_phone, target_email, role, expires_at, invited_by
        FROM merchant_staff_invitation
        WHERE invitation_code=? AND status=0 LIMIT 1`, in.InvitationCode)
	if err == sqlx.ErrNotFound {
		return nil, errors.New("invitation not found or already used")
	}
	if err != nil {
		return nil, err
	}
	if inv.ExpiresAt < time.Now().Unix() {
		_, _ = l.svcCtx.UserDB.ExecCtx(l.ctx,
			"UPDATE merchant_staff_invitation SET status=2 WHERE id=?", inv.Id)
		return nil, errors.New("invitation expired")
	}

	// 防代领：acceptor 的 phone/email 必须等于 invitation target
	if inv.TargetPhone != "" {
		if in.AcceptorPhone == "" || in.AcceptorPhone != inv.TargetPhone {
			return nil, errors.New("phone mismatch with invitation target")
		}
	} else if inv.TargetEmail != "" {
		if in.AcceptorEmail == "" || in.AcceptorEmail != inv.TargetEmail {
			return nil, errors.New("email mismatch with invitation target")
		}
	}

	// 防重复绑定（UNIQUE (shop_id, user_id) 也兜底）
	var existed int64
	_ = l.svcCtx.UserDB.QueryRowCtx(l.ctx, &existed,
		"SELECT COUNT(*) FROM merchant_staff WHERE shop_id=? AND user_id=?",
		inv.ShopId, in.AcceptorUid)
	if existed > 0 {
		return nil, errors.New("already a staff of this shop")
	}

	now := time.Now().Unix()
	if _, err := l.svcCtx.UserDB.ExecCtx(l.ctx, `
        INSERT INTO merchant_staff
          (shop_id, user_id, role, status, invited_by, joined_at)
        VALUES (?, ?, ?, 1, ?, ?)`,
		inv.ShopId, in.AcceptorUid, inv.Role, inv.InvitedBy, now); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.UserDB.ExecCtx(l.ctx,
		"UPDATE merchant_staff_invitation SET status=1, accepted_by=?, accepted_at=? WHERE id=?",
		in.AcceptorUid, now, inv.Id); err != nil {
		return nil, err
	}

	// 拿店名给 FE 展示
	var shopName string
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &shopName,
		"SELECT name FROM shop WHERE id=?", inv.ShopId)

	return &shop.AcceptInvitationResp{
		ShopId: int64(inv.ShopId), Role: inv.Role, ShopName: shopName,
	}, nil
}
