package logic

import (
	"context"
	"errors"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ChangePassword verifies the old password, checks the new one against the
// strength + history policy, then atomically rotates the hash and bumps the
// last_password_change column. Subject is `user` (1) or `admin` (2).
func (l *ChangePasswordLogic) ChangePassword(in *user.ChangePasswordReq) (*user.ChangePasswordResp, error) {
	if in.SubjectId <= 0 {
		return nil, errors.New("subject_id required")
	}
	oldPlain := trimPlain(in.OldPassword)
	newPlain := trimPlain(in.NewPassword)
	if oldPlain == "" || newPlain == "" {
		return nil, errors.New("old/new password required")
	}
	subjectType := int(in.SubjectType)
	if subjectType != subjectTypeUser && subjectType != subjectTypeAdmin {
		return nil, errors.New("subject_type must be 1 (user) or 2 (admin)")
	}

	var (
		hashCol  string
		table    string
	)
	if subjectType == subjectTypeAdmin {
		hashCol = "password_hash"
		table = "admin_user"
	} else {
		hashCol = "password"
		table = "`user`"
	}

	// 1. Load current hash.
	var currentHash string
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &currentHash,
		"SELECT "+hashCol+" FROM "+table+" WHERE id=? LIMIT 1", in.SubjectId); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("subject not found")
		}
		return nil, err
	}

	// 2. Verify old.
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPlain)); err != nil {
		return nil, errors.New("旧密码不正确")
	}

	// 3. Strength.
	if err := validatePassword(newPlain, l.svcCtx.PasswordPolicy); err != nil {
		return nil, err
	}

	// 4. Same-as-old short-circuit (avoids hashing twice for the common case).
	if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(newPlain)) == nil {
		return nil, errors.New("新密码不能与当前密码相同")
	}

	// 5. History.
	if err := passwordReusedRecently(l.ctx, l.svcCtx.DB, subjectType, uint64(in.SubjectId), newPlain, l.svcCtx.PasswordPolicy.MaxHistory); err != nil {
		return nil, err
	}

	// 6. Hash + persist.
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPlain), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE "+table+" SET "+hashCol+"=?, last_password_change=? WHERE id=?",
		string(newHash), now, in.SubjectId); err != nil {
		return nil, err
	}

	// 7. Record history (best-effort: a failure here doesn't roll back the
	// password change since the user already has a working new credential).
	if err := recordPasswordHistory(l.ctx, l.svcCtx.DB, subjectType, uint64(in.SubjectId), string(newHash), l.svcCtx.PasswordPolicy.MaxHistory); err != nil {
		l.Logger.Errorf("ChangePassword: recordPasswordHistory failed: %v", err)
	}

	return &user.ChangePasswordResp{Ok: true}, nil
}
