package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func intToStr(n int32) string { return strconv.Itoa(int(n)) }

// ReceiveCoupon 用户领券。
//
// 并发安全：
//   1. SELECT ... FOR UPDATE 锁 template 行，串行化同一模板的并发领取
//   2. received_count + total_count CAS 检查防超发
//   3. UNIQUE INDEX (user_id, template_id, status) 在每人限领=1 时硬约束防重领
//
// 校验：
//   - 模板存在 + 上架 + 领取窗口内
//   - 已领数 < 总数
//   - 用户已领 (状态 0/1/2) 数量 < per_user_limit
//
// 有效期计算：
//   - valid_type=0 固定日期: expire_time = template.valid_end
//   - valid_type=1 领后 N 天: expire_time = receive_time + valid_days*86400
func ReceiveCoupon(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ReceiveCouponReq) (*promotion.ReceiveCouponResp, error) {
	if in.UserId <= 0 || in.TemplateId <= 0 {
		return nil, errors.New("user_id/template_id required")
	}
	now := time.Now().Unix()
	var newId int64

	err := svcCtx.DB.TransactCtx(ctx, func(c context.Context, tx sqlx.Session) error {
		// 1) 锁 template 行
		var t struct {
			ShopId            int64 `db:"shop_id"`
			Status            int32 `db:"status"`
			TotalCount        int32 `db:"total_count"`
			ReceivedCount     int32 `db:"received_count"`
			PerUserLimit      int32 `db:"per_user_limit"`
			ValidType         int32 `db:"valid_type"`
			ValidDays         int32 `db:"valid_days"`
			ValidStart        int64 `db:"valid_start"`
			ValidEnd          int64 `db:"valid_end"`
			ReceiveStart      int64 `db:"receive_start"`
			ReceiveEnd        int64 `db:"receive_end"`
			IsNewUserOnly     int8  `db:"is_new_user_only"`
			NewUserWithinDays int32 `db:"new_user_within_days"`
		}
		if err := tx.QueryRowCtx(c, &t,
			"SELECT shop_id, status, total_count, received_count, per_user_limit, valid_type, valid_days, valid_start, valid_end, receive_start, receive_end, is_new_user_only, new_user_within_days FROM coupon_template WHERE id = ? FOR UPDATE",
			in.TemplateId,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("券模板不存在")
			}
			return err
		}
		if t.Status != 1 {
			return errors.New("券模板未上架")
		}
		if now < t.ReceiveStart || now >= t.ReceiveEnd {
			return errors.New("不在领取窗口内")
		}
		if t.ReceivedCount >= t.TotalCount {
			return errors.New("已被领完")
		}
		// S2.3 新人券校验
		if t.IsNewUserOnly == 1 {
			if in.UserRegisterTime <= 0 {
				return errors.New("新人券需要用户注册时间信息")
			}
			days := t.NewUserWithinDays
			if days <= 0 {
				days = 7
			}
			if in.UserRegisterTime+int64(days)*86400 < now {
				return errors.New("仅注册 " + intToStr(days) + " 天内的新用户可领此券")
			}
		}

		// 2) 校验每人限领
		var alreadyOwned int64
		if err := tx.QueryRowCtx(c, &alreadyOwned,
			"SELECT COUNT(*) FROM coupon WHERE user_id = ? AND template_id = ? AND status IN (0, 1, 2)",
			in.UserId, in.TemplateId,
		); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if alreadyOwned >= int64(t.PerUserLimit) {
			return errors.New("已达每人限领数量")
		}

		// 3) 算 expire_time
		var expireTime int64
		if t.ValidType == 0 {
			expireTime = t.ValidEnd
		} else {
			expireTime = now + int64(t.ValidDays)*86400
		}

		// 4) INSERT coupon
		res, err := tx.ExecCtx(c,
			"INSERT INTO coupon (template_id, user_id, shop_id, status, receive_time, expire_time) VALUES (?, ?, ?, 0, ?, ?)",
			in.TemplateId, in.UserId, t.ShopId, now, expireTime,
		)
		if err != nil {
			return err
		}
		newId, err = res.LastInsertId()
		if err != nil {
			return err
		}

		// 5) 更新 received_count
		if _, err := tx.ExecCtx(c,
			"UPDATE coupon_template SET received_count = received_count + 1 WHERE id = ? AND received_count < total_count",
			in.TemplateId,
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &promotion.ReceiveCouponResp{CouponId: newId}, nil
}
