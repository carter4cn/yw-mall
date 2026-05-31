// 券生命周期：锁定 / 释放 / 核销 —— 服务于下单 → 取消 → 支付成功 链路。
//
// 状态机：0 未用 -> 1 已锁定 -> 2 已使用
//                     |
//                     +--> 0 (订单取消/超时释放)
//                     +--> 3 (过期, by ListMyCoupons 副作用)
package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

// LockCoupon 下单时调，把 [coupon_ids] 的 status 由 0→1，绑定 order_id。
//
// SQL CAS：用 status=0 WHERE 条件 + 校验 RowsAffected = len(coupon_ids)。
// 若不匹配 (别的订单刚抢走 / 用户登错), 整体回滚返错。
func LockCoupon(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.LockCouponReq) (*promotion.OkResp, error) {
	if in.OrderId <= 0 || in.UserId <= 0 {
		return nil, errors.New("order_id/user_id required")
	}
	if len(in.CouponIds) == 0 {
		return &promotion.OkResp{Ok: true}, nil // 无券订单也允许
	}
	now := time.Now().Unix()
	placeholders := strings.Repeat("?,", len(in.CouponIds))
	placeholders = placeholders[:len(placeholders)-1]
	q := fmt.Sprintf(
		"UPDATE coupon SET status = 1, lock_time = ?, order_id = ? WHERE id IN (%s) AND user_id = ? AND status = 0 AND expire_time > ?",
		placeholders)
	args := []any{now, in.OrderId}
	for _, id := range in.CouponIds {
		args = append(args, id)
	}
	args = append(args, in.UserId, now)

	res, err := svcCtx.DB.ExecCtx(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected != int64(len(in.CouponIds)) {
		// 部分失败 → 回滚已锁定的
		_, _ = svcCtx.DB.ExecCtx(ctx,
			"UPDATE coupon SET status = 0, lock_time = 0, order_id = 0 WHERE order_id = ? AND status = 1",
			in.OrderId,
		)
		return nil, errors.New("券状态已变 (可能被其他订单使用或已过期)")
	}
	return &promotion.OkResp{Ok: true}, nil
}

// ReleaseCoupon 订单取消 / 5 分钟未付时调，把该订单所有锁定券归还到 0 状态。
// 幂等：不存在订单的券或已 consumed (status=2) 的不动。
func ReleaseCoupon(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ReleaseCouponReq) (*promotion.OkResp, error) {
	if in.OrderId <= 0 {
		return nil, errors.New("order_id required")
	}
	if _, err := svcCtx.DB.ExecCtx(ctx,
		"UPDATE coupon SET status = 0, lock_time = 0, order_id = 0 WHERE order_id = ? AND status = 1",
		in.OrderId,
	); err != nil {
		return nil, err
	}
	return &promotion.OkResp{Ok: true}, nil
}

// ConsumeCoupon 支付成功时调，把该订单的锁定券置为已使用 (1→2)。
// 同时 increment coupon_template.used_count。幂等：status=2 的不动。
func ConsumeCoupon(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.ConsumeCouponReq) (*promotion.OkResp, error) {
	if in.OrderId <= 0 {
		return nil, errors.New("order_id required")
	}
	now := time.Now().Unix()

	// 1) 先取锁定的券 template_ids (用于后续 increment used_count)
	var tplIds []int64
	if err := svcCtx.DB.QueryRowsCtx(ctx, &tplIds,
		"SELECT template_id FROM coupon WHERE order_id = ? AND status = 1", in.OrderId); err != nil {
		return nil, err
	}

	// 2) 状态 1 → 2 + use_time
	if _, err := svcCtx.DB.ExecCtx(ctx,
		"UPDATE coupon SET status = 2, use_time = ? WHERE order_id = ? AND status = 1",
		now, in.OrderId,
	); err != nil {
		return nil, err
	}

	// 3) increment used_count (best-effort, 失败不阻塞核销)
	for _, tid := range tplIds {
		_, _ = svcCtx.DB.ExecCtx(ctx,
			"UPDATE coupon_template SET used_count = used_count + 1 WHERE id = ?", tid,
		)
	}
	return &promotion.OkResp{Ok: true}, nil
}
