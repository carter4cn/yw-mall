package logic

import (
	"context"
	"errors"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPasswordByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordByCodeLogic {
	return &ResetPasswordByCodeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

const sceneResetPassword = 4

// ResetPasswordByCode 用 phone + 短信验证码重置密码（不验旧密码）。
// 走 scene=4，所以 FE 发码也要传 scene=4。
func (l *ResetPasswordByCodeLogic) ResetPasswordByCode(in *user.ResetPasswordByCodeReq) (*user.OkResp, error) {
	if !phoneRE.MatchString(in.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	newPlain := trimPlain(in.NewPassword)
	if err := validatePassword(newPlain, l.svcCtx.PasswordPolicy); err != nil {
		return nil, err
	}

	if err := consumeVerifyCode(l.ctx, l.svcCtx, sceneResetPassword, in.Phone,
		in.VerifyCode, in.ChallengeToken); err != nil {
		return nil, err
	}

	phoneHash := cryptox.Hmac(in.Phone)
	var uid uint64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &uid,
		"SELECT id FROM `user` WHERE phone_hash=? LIMIT 1", phoneHash)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("该手机号未注册")
		}
		return nil, err
	}

	if err := passwordReusedRecently(l.ctx, l.svcCtx.DB, subjectTypeUser,
		uid, newPlain, l.svcCtx.PasswordPolicy.MaxHistory); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPlain), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `user` SET password=?, last_password_change=? WHERE id=?",
		string(hash), now, uid); err != nil {
		return nil, err
	}

	_ = recordPasswordHistory(l.ctx, l.svcCtx.DB, subjectTypeUser, uid,
		string(hash), l.svcCtx.PasswordPolicy.MaxHistory)

	// Force-logout: pre-reset stolen tokens must not survive. Only "user" role —
	// admin path can't reach here, but the index is shared, so role-filter is
	// the safe move (see destroyUserSessionsByRole comment).
	if n, derr := destroyUserSessionsByRole(l.ctx, l.svcCtx, int64(uid), "user"); derr != nil {
		l.Logger.Errorf("ResetPasswordByCode: destroyUserSessionsByRole(uid=%d) failed: %v", uid, derr)
	} else if n > 0 {
		l.Logger.Infof("ResetPasswordByCode: kicked %d user session(s) for uid=%d", n, uid)
	}

	return &user.OkResp{Ok: true}, nil
}
