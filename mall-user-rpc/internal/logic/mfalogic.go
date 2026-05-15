package logic

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MfaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MfaLogic {
	return &MfaLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

const (
	mfaIssuer       = "yw-mall-admin"
	backupCodeCount = 5
	backupCodeLen   = 10
)

// Enable generates a fresh TOTP secret + 5 single-use backup codes for the
// admin and stores them with enabled=0. The admin must call Confirm with a
// valid code from their authenticator app to set enabled=1.
func (l *MfaLogic) Enable(in *user.EnableAdminMfaReq) (*user.EnableAdminMfaResp, error) {
	if in.AdminId <= 0 {
		return nil, errors.New("admin_id required")
	}

	// Lookup the admin's username to embed in the otpauth URL (so the user's
	// authenticator app shows e.g. "yw-mall-admin (alice)").
	var username string
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &username,
		"SELECT username FROM admin_user WHERE id=? LIMIT 1", in.AdminId); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      mfaIssuer,
		AccountName: username,
		Period:      30,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, fmt.Errorf("totp.Generate: %w", err)
	}

	secret := key.Secret()
	secretEnc, err := cryptox.Encrypt(secret)
	if err != nil {
		return nil, err
	}

	backupCodes, err := generateBackupCodes(backupCodeCount, backupCodeLen)
	if err != nil {
		return nil, err
	}
	backupCodesEnc, err := cryptox.Encrypt(strings.Join(backupCodes, ","))
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `
		INSERT INTO admin_mfa (admin_id, totp_secret_enc, backup_codes_enc, enabled, created_at, last_used_at)
		VALUES (?, ?, ?, 0, ?, 0)
		ON DUPLICATE KEY UPDATE
		  totp_secret_enc=VALUES(totp_secret_enc),
		  backup_codes_enc=VALUES(backup_codes_enc),
		  enabled=0,
		  created_at=VALUES(created_at)`,
		in.AdminId, secretEnc, backupCodesEnc, now); err != nil {
		return nil, err
	}

	return &user.EnableAdminMfaResp{
		TotpSecret:  secret,
		QrUrl:       key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// Confirm verifies the code against the secret and flips enabled=1.
func (l *MfaLogic) Confirm(in *user.ConfirmAdminMfaReq) (*user.OkResp, error) {
	secret, _, err := l.loadSecret(in.AdminId)
	if err != nil {
		return nil, err
	}
	if !totp.Validate(in.Code, secret) {
		return nil, errors.New("MFA 验证码不正确")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE admin_mfa SET enabled=1, last_used_at=? WHERE admin_id=?",
		now, in.AdminId); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}

// Verify checks the code (TOTP first, backup-code fallback) without changing
// the enabled flag. Used at login time after the password check passed.
func (l *MfaLogic) Verify(in *user.VerifyAdminMfaReq) (*user.OkResp, error) {
	secret, backupCodes, err := l.loadSecret(in.AdminId)
	if err != nil {
		return nil, err
	}
	if totp.Validate(in.Code, secret) {
		_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
			"UPDATE admin_mfa SET last_used_at=? WHERE admin_id=?",
			time.Now().Unix(), in.AdminId)
		return &user.OkResp{Ok: true}, nil
	}
	// Backup-code fallback. One-shot: matched code is removed.
	for i, bc := range backupCodes {
		if bc == in.Code {
			remaining := append([]string{}, backupCodes[:i]...)
			remaining = append(remaining, backupCodes[i+1:]...)
			joined := strings.Join(remaining, ",")
			enc, err := cryptox.Encrypt(joined)
			if err != nil {
				return nil, err
			}
			_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
				"UPDATE admin_mfa SET backup_codes_enc=?, last_used_at=? WHERE admin_id=?",
				enc, time.Now().Unix(), in.AdminId)
			return &user.OkResp{Ok: true}, nil
		}
	}
	return nil, errors.New("MFA 验证码不正确")
}

// Disable verifies one final TOTP, then removes the row entirely so a fresh
// Enable later starts from a clean slate.
func (l *MfaLogic) Disable(in *user.DisableAdminMfaReq) (*user.OkResp, error) {
	secret, _, err := l.loadSecret(in.AdminId)
	if err != nil {
		return nil, err
	}
	if !totp.Validate(in.Code, secret) {
		return nil, errors.New("MFA 验证码不正确")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"DELETE FROM admin_mfa WHERE admin_id=?", in.AdminId); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}

// Status returns enabled+last_used. Returns enabled=false when no row exists
// (so the admin-fe doesn't need to special-case "not yet provisioned").
func (l *MfaLogic) Status(in *user.GetAdminMfaStatusReq) (*user.GetAdminMfaStatusResp, error) {
	var row struct {
		Enabled    int32 `db:"enabled"`
		LastUsedAt int64 `db:"last_used_at"`
	}
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT enabled, last_used_at FROM admin_mfa WHERE admin_id=? LIMIT 1", in.AdminId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &user.GetAdminMfaStatusResp{Enabled: false}, nil
		}
		return nil, err
	}
	return &user.GetAdminMfaStatusResp{Enabled: row.Enabled == 1, LastUsedAt: row.LastUsedAt}, nil
}

// loadSecret returns the decrypted TOTP secret + the (possibly empty) list of
// remaining backup codes for the admin.
func (l *MfaLogic) loadSecret(adminId int64) (string, []string, error) {
	if adminId <= 0 {
		return "", nil, errors.New("admin_id required")
	}
	var row struct {
		SecretEnc string `db:"totp_secret_enc"`
		BackupEnc string `db:"backup_codes_enc"`
	}
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT totp_secret_enc, COALESCE(backup_codes_enc,'') AS backup_codes_enc FROM admin_mfa WHERE admin_id=? LIMIT 1",
		adminId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return "", nil, errors.New("MFA 未配置")
		}
		return "", nil, err
	}
	secret, err := cryptox.Decrypt(row.SecretEnc)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt totp secret: %w", err)
	}
	var backupCodes []string
	if row.BackupEnc != "" {
		bc, err := cryptox.DecryptIfCiphertext(row.BackupEnc)
		if err != nil {
			return "", nil, fmt.Errorf("decrypt backup codes: %w", err)
		}
		if bc != "" {
			backupCodes = strings.Split(bc, ",")
		}
	}
	return secret, backupCodes, nil
}

// generateBackupCodes builds N codes of length L using crypto/rand → base32.
// Base32 keeps them human-typable (no 0/O/I/1 confusion) and short.
func generateBackupCodes(n, l int) ([]string, error) {
	out := make([]string, 0, n)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	for i := 0; i < n; i++ {
		buf := make([]byte, (l*5)/8+1)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		s := enc.EncodeToString(buf)
		if len(s) > l {
			s = s[:l]
		}
		out = append(out, strings.ToUpper(s))
	}
	return out, nil
}
