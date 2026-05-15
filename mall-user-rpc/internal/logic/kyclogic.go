package logic

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type KycLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKycLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KycLogic {
	return &KycLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// KYC status enum (mirrored on the FE for badges).
const (
	kycStatusUnsubmitted = 0
	kycStatusReviewing   = 1
	kycStatusPassed      = 2
	kycStatusRejected    = 3
)

// Submit writes/overwrites the row with status=1 (审核中) and kicks the mock
// provider in a goroutine so the user response is immediate.
func (l *KycLogic) Submit(in *user.SubmitKycReq) (*user.SubmitKycResp, error) {
	if in.UserId <= 0 || in.RealName == "" || in.IdCardNo == "" {
		return nil, errors.New("缺少必填字段")
	}
	realNameEnc, err := cryptox.Encrypt(strings.TrimSpace(in.RealName))
	if err != nil {
		return nil, err
	}
	idCardEnc, err := cryptox.Encrypt(strings.TrimSpace(in.IdCardNo))
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	requestId := uuid.NewString()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `
		INSERT INTO user_kyc (user_id, status, real_name_enc, id_card_no_enc, id_card_front_url, id_card_back_url, face_video_url, submit_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  status=?, real_name_enc=VALUES(real_name_enc), id_card_no_enc=VALUES(id_card_no_enc),
		  id_card_front_url=VALUES(id_card_front_url), id_card_back_url=VALUES(id_card_back_url),
		  face_video_url=VALUES(face_video_url), reject_reason='', submit_time=VALUES(submit_time),
		  audit_time=0, audit_admin_id=0`,
		in.UserId, kycStatusReviewing, realNameEnc, idCardEnc,
		in.IdCardFrontUrl, in.IdCardBackUrl, in.FaceVideoUrl, now,
		kycStatusReviewing); err != nil {
		return nil, err
	}

	// Kick mock provider asynchronously: random 90% pass, 500ms delay.
	go l.runMockAudit(in.UserId)

	return &user.SubmitKycResp{RequestId: requestId, Status: kycStatusReviewing}, nil
}

// Status returns the current KYC row decrypted. Empty row → status=0.
func (l *KycLogic) Status(in *user.GetKycStatusReq) (*user.GetKycStatusResp, error) {
	var row struct {
		Status       int32  `db:"status"`
		RealNameEnc  string `db:"real_name_enc"`
		IdCardNoEnc  string `db:"id_card_no_enc"`
		RejectReason string `db:"reject_reason"`
		SubmitTime   int64  `db:"submit_time"`
		AuditTime    int64  `db:"audit_time"`
	}
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row, `
		SELECT status, real_name_enc, id_card_no_enc, reject_reason, submit_time, audit_time
		FROM user_kyc WHERE user_id=? LIMIT 1`, in.UserId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &user.GetKycStatusResp{Status: kycStatusUnsubmitted}, nil
		}
		return nil, err
	}
	realName, _ := cryptox.DecryptIfCiphertext(row.RealNameEnc)
	idCardNo, _ := cryptox.DecryptIfCiphertext(row.IdCardNoEnc)
	return &user.GetKycStatusResp{
		Status:       row.Status,
		RejectReason: row.RejectReason,
		SubmitTime:   row.SubmitTime,
		AuditTime:    row.AuditTime,
		RealName:     realName,
		IdCardNo:     idCardNo,
	}, nil
}

// ListPending serves the admin review queue. Includes status=1 (审核中) only.
func (l *KycLogic) ListPending(in *user.ListPendingKycReq) (*user.ListPendingKycResp, error) {
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM user_kyc WHERE status=?", kycStatusReviewing); err != nil {
		return nil, err
	}

	type row struct {
		UserId         uint64 `db:"user_id"`
		Status         int32  `db:"status"`
		RealNameEnc    string `db:"real_name_enc"`
		IdCardNoEnc    string `db:"id_card_no_enc"`
		IdCardFrontUrl string `db:"id_card_front_url"`
		IdCardBackUrl  string `db:"id_card_back_url"`
		FaceVideoUrl   string `db:"face_video_url"`
		SubmitTime     int64  `db:"submit_time"`
		Username       string `db:"username"`
	}
	var rows []*row
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, `
		SELECT k.user_id, k.status, k.real_name_enc, k.id_card_no_enc,
		       k.id_card_front_url, k.id_card_back_url, k.face_video_url,
		       k.submit_time, COALESCE(u.username,'') AS username
		FROM user_kyc k LEFT JOIN `+"`user`"+` u ON u.id = k.user_id
		WHERE k.status=? ORDER BY k.submit_time ASC LIMIT ? OFFSET ?`,
		kycStatusReviewing, pageSize, offset); err != nil {
		return nil, err
	}

	out := make([]*user.KycPendingItem, 0, len(rows))
	for _, r := range rows {
		realName, _ := cryptox.DecryptIfCiphertext(r.RealNameEnc)
		idCardNo, _ := cryptox.DecryptIfCiphertext(r.IdCardNoEnc)
		out = append(out, &user.KycPendingItem{
			UserId:         int64(r.UserId),
			Username:       r.Username,
			RealName:       realName,
			IdCardNo:       idCardNo,
			IdCardFrontUrl: r.IdCardFrontUrl,
			IdCardBackUrl:  r.IdCardBackUrl,
			FaceVideoUrl:   r.FaceVideoUrl,
			SubmitTime:     r.SubmitTime,
			Status:         r.Status,
		})
	}
	return &user.ListPendingKycResp{Items: out, Total: total}, nil
}

// AdminAudit lets a human override the mock provider's verdict.
func (l *KycLogic) AdminAudit(in *user.AdminAuditKycReq) (*user.OkResp, error) {
	if in.UserId <= 0 || in.AuditAdminId <= 0 {
		return nil, errors.New("user_id 和 audit_admin_id 必填")
	}
	status := kycStatusRejected
	if in.Pass {
		status = kycStatusPassed
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `
		UPDATE user_kyc SET status=?, reject_reason=?, audit_time=?, audit_admin_id=?
		WHERE user_id=?`,
		status, in.Reason, time.Now().Unix(), in.AuditAdminId, in.UserId); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}

// runMockAudit is the in-process mock for an external IDV provider. 500ms
// delay then 90% pass rate. We use a fresh background context because the
// caller's ctx will already be cancelled once the RPC response goes out.
func (l *KycLogic) runMockAudit(userId int64) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("KYC mock audit panicked for user=%d: %v", userId, r)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	pass := rand.Intn(10) != 0 // 90% pass
	status := kycStatusPassed
	reason := ""
	if !pass {
		status = kycStatusRejected
		reason = "mock provider: face match failed"
	}
	bg := context.Background()
	if _, err := l.svcCtx.DB.ExecCtx(bg, `
		UPDATE user_kyc SET status=?, reject_reason=?, audit_time=?, audit_admin_id=0
		WHERE user_id=? AND status=?`,
		status, reason, time.Now().Unix(), userId, kycStatusReviewing); err != nil {
		logx.Errorf("KYC mock audit update failed for user=%d: %v", userId, err)
	}
	logx.Infof("[mock-kyc] user=%d %s", userId, fmt.Sprintf("status=%d reason=%q", status, reason))
}
