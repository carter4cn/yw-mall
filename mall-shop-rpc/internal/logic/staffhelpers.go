package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"

	"mall-shop-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 邀请 token 有效期 7 天
const invitationTTL = 7 * 24 * 3600

// 复用 reg-v2 同款 phone/email regex（避免跨 service import）
var (
	invPhoneRE = regexp.MustCompile(`^\d{11}$`)
	invEmailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// staffRow 是 merchant_staff 行的最小子集
type staffRow struct {
	Id       uint64 `db:"id"`
	ShopId   uint64 `db:"shop_id"`
	UserId   uint64 `db:"user_id"`
	Role     string `db:"role"`
	Status   int32  `db:"status"`
	JoinedAt int64  `db:"joined_at"`
}

// queryStaffByUserId 找当前用户绑定的（active）staff 记录。
// 没绑或被冻结返 nil, nil。
func queryStaffByUserId(ctx context.Context, svcCtx *svc.ServiceContext, uid int64) (*staffRow, error) {
	var s staffRow
	err := svcCtx.UserDB.QueryRowCtx(ctx, &s, `
        SELECT id, shop_id, user_id, role, status, joined_at
        FROM merchant_staff WHERE user_id=? AND status=1 LIMIT 1`, uid)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// requireShopOwner 校验 uid 是 shopId 的 owner，否则返业务错。
func requireShopOwner(ctx context.Context, svcCtx *svc.ServiceContext, shopId, uid int64) error {
	if shopId <= 0 || uid <= 0 {
		return errors.New("invalid shop or uid")
	}
	var role string
	err := svcCtx.UserDB.QueryRowCtx(ctx, &role, `
        SELECT role FROM merchant_staff
        WHERE shop_id=? AND user_id=? AND status=1 LIMIT 1`, shopId, uid)
	if err != nil {
		return errors.New("not a staff of this shop")
	}
	if role != RoleOwner {
		return errors.New("only shop owner can perform this action")
	}
	return nil
}

// newInvitationCode 生成 32 字节 URL-safe base64 编码的不可猜 token
func newInvitationCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
