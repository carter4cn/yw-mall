package logic

import (
	"context"
	"errors"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CreateCouponTemplate 商家创建券模板。
// 同时在 activity 表插一行 type=coupon 占位，让券与活动同表统计。
func CreateCouponTemplate(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.CreateCouponTemplateReq) (*promotion.CreateCouponTemplateResp, error) {
	if in.Name == "" || in.Type == "" || in.Value <= 0 || in.TotalCount <= 0 {
		return nil, errors.New("name/type/value/total_count required")
	}
	if in.ValidType == 0 && (in.ValidStart == 0 || in.ValidEnd == 0) {
		return nil, errors.New("固定日期模式需填 valid_start/valid_end")
	}
	if in.ValidType == 1 && in.ValidDays <= 0 {
		return nil, errors.New("领取后 N 天模式需填 valid_days")
	}
	if in.ReceiveStart == 0 || in.ReceiveEnd <= in.ReceiveStart {
		return nil, errors.New("invalid receive window")
	}
	perUserLimit := in.PerUserLimit
	if perUserLimit <= 0 {
		perUserLimit = 1
	}

	var newId int64
	err := svcCtx.DB.TransactCtx(ctx, func(c context.Context, tx sqlx.Session) error {
		// 1) 创建占位 activity (type=coupon)
		res, err := tx.ExecCtx(c,
			`INSERT INTO activity (type, name, shop_id, status, start_time, end_time, priority, stackable, description)
			 VALUES ('coupon', ?, ?, 2, ?, ?, 0, 1, ?)`,
			in.Name, in.ShopId, in.ReceiveStart, in.ReceiveEnd, "券模板: "+in.Name,
		)
		if err != nil {
			return err
		}
		activityId, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// 2) 创建 coupon_template
		newUserDays := in.NewUserWithinDays
		if in.IsNewUserOnly && newUserDays <= 0 {
			newUserDays = 7 // 默认 7 天
		}
		res, err = tx.ExecCtx(c,
			`INSERT INTO coupon_template
			   (activity_id, shop_id, name, type, value, min_amount, max_discount, category_id,
			    total_count, per_user_limit, valid_type, valid_days, valid_start, valid_end,
			    receive_start, receive_end, status, is_new_user_only, new_user_within_days)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
			activityId, in.ShopId, in.Name, in.Type, in.Value, in.MinAmount, in.MaxDiscount, in.CategoryId,
			in.TotalCount, perUserLimit, in.ValidType, in.ValidDays, in.ValidStart, in.ValidEnd,
			in.ReceiveStart, in.ReceiveEnd, boolToTinyint(in.IsNewUserOnly), newUserDays,
		)
		if err != nil {
			return err
		}
		newId, err = res.LastInsertId()
		return err
	})
	if err != nil {
		return nil, err
	}
	return &promotion.CreateCouponTemplateResp{Id: newId}, nil
}
