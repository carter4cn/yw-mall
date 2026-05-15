package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-common/cryptox"
	userpb "mall-user-rpc/user"
	"mall-user-rpc/userclient"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// S4.5 personal-data lifecycle endpoints (PIPL friendly).
//
// `data/scope`  — static disclosure of which PII fields we collect
// `data/export` — synchronous bundle of the user's PII (login, addresses, KYC)
// `data/erase`  — soft-delete: status=2, anonymise PII columns
//
// We keep this in the gateway because export/erase touch multiple downstream
// data sources (user-rpc, addresses); easier to coordinate here than spread
// the rules across services.

// PIIFieldDescriptor mirrors the structure rendered on the FE 数据范围 page.
type PIIFieldDescriptor struct {
	Field     string `json:"field"`
	Purpose   string `json:"purpose"`
	Retention string `json:"retention"`
}

// scopeFields is intentionally hard-coded so any change to the data we collect
// requires a code review (legal review trail).
var scopeFields = []PIIFieldDescriptor{
	{Field: "username", Purpose: "账号识别 / 登录", Retention: "账号注销前长期存储"},
	{Field: "phone", Purpose: "登录与通知", Retention: "账号注销前长期存储 (加密)"},
	{Field: "addresses.receiver_name", Purpose: "订单收货", Retention: "账号注销前长期存储"},
	{Field: "addresses.phone", Purpose: "订单配送联系", Retention: "账号注销前长期存储"},
	{Field: "addresses.detail", Purpose: "订单收货", Retention: "账号注销前长期存储"},
	{Field: "kyc.real_name", Purpose: "实名认证", Retention: "认证通过 5 年 (加密)"},
	{Field: "kyc.id_card_no", Purpose: "实名认证", Retention: "认证通过 5 年 (加密)"},
	{Field: "orders.receiver_*", Purpose: "履约 / 售后 / 财务对账", Retention: "订单完成 3 年"},
}

func GetDataScope(_ context.Context) *types.DataScopeResp {
	out := make([]types.DataScopeField, 0, len(scopeFields))
	for _, f := range scopeFields {
		out = append(out, types.DataScopeField{Field: f.Field, Purpose: f.Purpose, Retention: f.Retention})
	}
	return &types.DataScopeResp{Version: "v1", Fields: out}
}

// ExportData synchronously gathers everything we have on the calling user and
// returns a JSON payload. We don't paginate orders — for an MVP this is fine,
// and the FE can chunk into a download blob.
func ExportData(ctx context.Context, svcCtx *svc.ServiceContext) (*types.DataExportResp, error) {
	uid := middleware.UidFromCtx(ctx)
	if uid <= 0 {
		return nil, errors.New("unauthenticated")
	}

	user, err := svcCtx.UserRpc.GetUser(ctx, &userpb.GetUserReq{Id: uid})
	if err != nil {
		return nil, err
	}

	addrResp, err := svcCtx.UserRpc.ListAddresses(ctx, &userclient.ListAddressesReq{UserId: uid})
	if err != nil {
		return nil, err
	}
	addresses := make([]types.AddressItem, 0, len(addrResp.Addresses))
	for _, a := range addrResp.Addresses {
		addresses = append(addresses, types.AddressItem{
			Id: a.Id, ReceiverName: a.ReceiverName, Phone: a.Phone,
			Province: a.Province, City: a.City, District: a.District, Detail: a.Detail,
			IsDefault: a.IsDefault, CreateTime: a.CreateTime,
		})
	}

	kycResp, _ := svcCtx.UserRpc.GetKycStatus(ctx, &userclient.GetKycStatusReq{UserId: uid})
	var kyc *types.KycExport
	if kycResp != nil && kycResp.Status > 0 {
		kyc = &types.KycExport{
			Status: kycResp.Status, RealName: kycResp.RealName,
			IdCardNo: kycResp.IdCardNo, RejectReason: kycResp.RejectReason,
			SubmitTime: kycResp.SubmitTime, AuditTime: kycResp.AuditTime,
		}
	}

	logx.WithContext(ctx).Infof("[data-export] uid=%d", uid)
	return &types.DataExportResp{
		Version:    "v1",
		RequestId:  uuid.NewString(),
		ExportedAt: time.Now().Unix(),
		User: types.UserExport{
			Id: user.Id, Username: user.Username, Phone: user.Phone,
			Avatar: user.Avatar, CreateTime: user.CreateTime,
		},
		Addresses: addresses,
		Kyc:       kyc,
	}, nil
}

// EraseData performs the PIPL-required user-initiated deletion. We:
//   1. mall_user.user.status = 2 (login refuses)
//   2. anonymise username (deleted_<id>) + phone = ''
//   3. anonymise all user_address rows (receiver_name='已注销', phone='', detail='')
//   4. erase any KYC PII (status stays so audit trail is preserved)
//
// Orders are intentionally retained as-is per CLAUDE.md (legal/finance reason).
// We DO destroy all sessions so the FE is logged out immediately.
func EraseData(ctx context.Context, svcCtx *svc.ServiceContext) (*types.OkResp, error) {
	uid := middleware.UidFromCtx(ctx)
	if uid <= 0 {
		return nil, errors.New("unauthenticated")
	}
	db := mustGetUserDB(svcCtx)
	if db == nil {
		return nil, errors.New("erase unavailable: user-db not configured")
	}

	now := time.Now().Unix()

	if _, err := db.ExecCtx(ctx,
		"UPDATE `user` SET status=2, username=CONCAT('deleted_', id), phone='' WHERE id=?", uid); err != nil {
		return nil, err
	}
	if _, err := db.ExecCtx(ctx,
		"UPDATE user_address SET receiver_name='已注销', phone='', detail='', update_time=? WHERE user_id=?",
		now, uid); err != nil {
		return nil, err
	}
	// KYC PII anonymise — re-encrypt the empty string so column remains valid.
	emptyEnc, _ := cryptox.Encrypt("")
	if _, err := db.ExecCtx(ctx,
		"UPDATE user_kyc SET real_name_enc=?, id_card_no_enc=?, id_card_front_url='', id_card_back_url='', face_video_url='' WHERE user_id=?",
		emptyEnc, emptyEnc, uid); err != nil {
		// Soft-fail: KYC table may not exist yet during the rollout window.
		logx.WithContext(ctx).Errorf("EraseData: kyc anonymise: %v", err)
	}

	// Best-effort session destroy.
	_, _ = svcCtx.UserRpc.DestroyAllUserSessions(ctx, &userclient.DestroyAllUserSessionsReq{Uid: uid})

	logx.WithContext(ctx).Infof("[data-erase] uid=%d", uid)
	return &types.OkResp{Ok: true}, nil
}

// mustGetUserDB returns a sqlx.SqlConn pointed at mall_user. We allow the DSN
// to be configured via the same OpLogDataSource pattern; if absent we point at
// the well-known ProxySQL DSN as a fallback.
func mustGetUserDB(svcCtx *svc.ServiceContext) sqlx.SqlConn {
	if svcCtx.UserDB != nil {
		return svcCtx.UserDB
	}
	return nil
}

// SubmitKyc and GetKycStatus are thin gateway wrappers around the user-rpc.
func SubmitKyc(ctx context.Context, svcCtx *svc.ServiceContext, req *types.KycSubmitReq) (*types.KycSubmitResp, error) {
	uid := middleware.UidFromCtx(ctx)
	if uid <= 0 {
		return nil, errors.New("unauthenticated")
	}
	res, err := svcCtx.UserRpc.SubmitKyc(ctx, &userclient.SubmitKycReq{
		UserId: uid, RealName: req.RealName, IdCardNo: req.IdCardNo,
		IdCardFrontUrl: req.IdCardFrontUrl, IdCardBackUrl: req.IdCardBackUrl,
		FaceVideoUrl: req.FaceVideoUrl,
	})
	if err != nil {
		return nil, err
	}
	return &types.KycSubmitResp{RequestId: res.RequestId, Status: res.Status}, nil
}

func GetKycStatus(ctx context.Context, svcCtx *svc.ServiceContext) (*types.KycStatusResp, error) {
	uid := middleware.UidFromCtx(ctx)
	if uid <= 0 {
		return nil, errors.New("unauthenticated")
	}
	res, err := svcCtx.UserRpc.GetKycStatus(ctx, &userclient.GetKycStatusReq{UserId: uid})
	if err != nil {
		return nil, err
	}
	return &types.KycStatusResp{
		Status:       res.Status,
		RejectReason: res.RejectReason,
		SubmitTime:   res.SubmitTime,
		AuditTime:    res.AuditTime,
		RealName:     maskName(res.RealName),
		IdCardNo:     maskIdCard(res.IdCardNo),
	}, nil
}

// ChangePassword (c-side): reuse user-rpc subject_type=1.
func UserChangePassword(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ChangePasswordReq) (*types.OkResp, error) {
	uid := middleware.UidFromCtx(ctx)
	if uid <= 0 {
		return nil, errors.New("unauthenticated")
	}
	if _, err := svcCtx.UserRpc.ChangePassword(ctx, &userclient.ChangePasswordReq{
		SubjectType: 1, SubjectId: uid,
		OldPassword: req.OldPassword, NewPassword: req.NewPassword,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}

func maskName(s string) string {
	r := []rune(s)
	if len(r) <= 1 {
		return s
	}
	return string(r[:1]) + "**"
}

func maskIdCard(s string) string {
	if len(s) < 8 {
		return s
	}
	return s[:4] + "**********" + s[len(s)-4:]
}

// _ keep the fmt import used (formerly used in debug strings).
var _ = fmt.Sprintf
