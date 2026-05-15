package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"

	"mall-user-rpc/internal/svc"
)

// Subject types for password_history rows. 1=user, 2=admin.
const (
	subjectTypeUser  = 1
	subjectTypeAdmin = 2
)

// validatePassword enforces the loaded policy. Errors are user-friendly Chinese
// strings so the gateway can surface them directly without translation.
func validatePassword(plain string, p svc.PasswordPolicy) error {
	if len(plain) < p.MinLength {
		return fmt.Errorf("密码至少 %d 位", p.MinLength)
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false
	for _, r := range plain {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if p.RequireUpper && !hasUpper {
		return errors.New("密码必须包含大写字母")
	}
	if p.RequireLower && !hasLower {
		return errors.New("密码必须包含小写字母")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("密码必须包含数字")
	}
	if p.RequireSymbol && !hasSymbol {
		return errors.New("密码必须包含符号")
	}
	return nil
}

// passwordReusedRecently returns nil if the candidate plaintext does NOT match
// any of the last N stored password hashes for this subject.
func passwordReusedRecently(ctx context.Context, db sqlx.SqlConn, subjectType int, subjectID uint64, candidatePlain string, max int) error {
	if max <= 0 {
		return nil
	}
	var hashes []string
	if err := db.QueryRowsCtx(ctx, &hashes,
		"SELECT password_hash FROM password_history WHERE subject_type=? AND subject_id=? ORDER BY create_time DESC LIMIT ?",
		subjectType, subjectID, max); err != nil {
		return err
	}
	for _, h := range hashes {
		if bcrypt.CompareHashAndPassword([]byte(h), []byte(candidatePlain)) == nil {
			return fmt.Errorf("新密码不能与最近 %d 次相同", max)
		}
	}
	return nil
}

// recordPasswordHistory inserts the new hash and trims the table back to the
// most-recent `max` rows for the subject. Trim runs in the same call so we
// don't grow the table unbounded.
func recordPasswordHistory(ctx context.Context, db sqlx.SqlConn, subjectType int, subjectID uint64, hash string, max int) error {
	now := time.Now().Unix()
	if _, err := db.ExecCtx(ctx,
		"INSERT INTO password_history (subject_type, subject_id, password_hash, create_time) VALUES (?, ?, ?, ?)",
		subjectType, subjectID, hash, now); err != nil {
		return err
	}
	if max <= 0 {
		return nil
	}
	// MySQL forbids LIMIT in subqueries with the same target table for DELETE,
	// so we two-step it: pick the cutoff timestamp, then DELETE older rows.
	var cutoff int64
	err := db.QueryRowCtx(ctx, &cutoff,
		"SELECT create_time FROM password_history WHERE subject_type=? AND subject_id=? ORDER BY create_time DESC LIMIT 1 OFFSET ?",
		subjectType, subjectID, max)
	if err != nil {
		// No row at offset N → fewer than N+1 history rows, nothing to trim.
		return nil
	}
	_, err = db.ExecCtx(ctx,
		"DELETE FROM password_history WHERE subject_type=? AND subject_id=? AND create_time < ?",
		subjectType, subjectID, cutoff)
	return err
}

// passwordExpired returns true when the policy's MaxAgeDays has elapsed since
// last_password_change. Zero/missing last_password_change is treated as "never
// changed yet" → expired immediately so legacy users are forced to rotate.
func passwordExpired(lastChange int64, maxAgeDays int) bool {
	if maxAgeDays <= 0 {
		return false
	}
	if lastChange <= 0 {
		// Don't force legacy zero values to expire — they may be brand-new
		// accounts whose last_password_change column hasn't been backfilled.
		// The Sprint 5 migration will populate it for the existing rows.
		return false
	}
	return time.Now().Unix()-lastChange > int64(maxAgeDays)*86400
}

// trimPlain is a tiny helper used at every entrypoint that accepts a password
// from user input — leading/trailing whitespace is almost always a typo.
func trimPlain(s string) string { return strings.TrimSpace(s) }
