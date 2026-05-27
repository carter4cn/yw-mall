package logic

import (
	"context"
	"errors"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateInvitationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInvitationLogic {
	return &CreateInvitationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// 单店员工 + pending 邀请软上限
const maxStaffPerShop = 20

// CreateInvitation 仅 owner 可调。生成 32 字节不可猜 code，7 天过期，
// 通过 logx mock SMS/邮件下发（与 reg-v2 同模式）。
func (l *CreateInvitationLogic) CreateInvitation(in *shop.CreateInvitationReq) (*shop.CreateInvitationResp, error) {
	if !IsValidRole(in.Role) || in.Role == RoleOwner {
		return nil, errors.New("invalid role; owner cannot be invited")
	}
	if in.TargetPhone == "" && in.TargetEmail == "" {
		return nil, errors.New("target phone or email required")
	}
	if in.TargetPhone != "" && !invPhoneRE.MatchString(in.TargetPhone) {
		return nil, errors.New("invalid phone")
	}
	if in.TargetEmail != "" && !invEmailRE.MatchString(in.TargetEmail) {
		return nil, errors.New("invalid email")
	}
	if err := requireShopOwner(l.ctx, l.svcCtx, in.ShopId, in.InvitedByUid); err != nil {
		return nil, err
	}

	// 软上限：active staff + pending 邀请 ≤ 20
	var cnt int64
	_ = l.svcCtx.UserDB.QueryRowCtx(l.ctx, &cnt, `
        SELECT
          (SELECT COUNT(*) FROM merchant_staff WHERE shop_id=? AND status=1)
        + (SELECT COUNT(*) FROM merchant_staff_invitation
           WHERE shop_id=? AND status=0 AND expires_at>?)`,
		in.ShopId, in.ShopId, time.Now().Unix())
	if cnt >= maxStaffPerShop {
		return nil, errors.New("staff cap reached (20)")
	}

	code, err := newInvitationCode()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	expires := now + invitationTTL
	_, err = l.svcCtx.UserDB.ExecCtx(l.ctx, `
        INSERT INTO merchant_staff_invitation
          (shop_id, invited_by, target_phone, target_email,
           role, invitation_code, status, expires_at)
        VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
		in.ShopId, in.InvitedByUid, in.TargetPhone, in.TargetEmail,
		in.Role, code, expires)
	if err != nil {
		return nil, err
	}

	// mock SMS / 邮件下发，与 reg-v2 SendVerifyCode 同模式
	target := in.TargetPhone
	if target == "" {
		target = in.TargetEmail
	}
	logx.WithContext(l.ctx).Infof("[mock-invite] target=%s code=%s role=%s shop_id=%d",
		target, code, in.Role, in.ShopId)

	return &shop.CreateInvitationResp{InvitationCode: code, ExpiresAt: expires}, nil
}
