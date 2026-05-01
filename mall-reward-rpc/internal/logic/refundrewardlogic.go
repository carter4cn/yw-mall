package logic

import (
	"context"
	"fmt"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RefundRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundRewardLogic {
	return &RefundRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefundReward is the SAGA compensation hook. It flips the record to REFUNDED
// and short-circuits any still-PENDING outbox row (so the relay won't publish
// a now-cancelled dispatch). Logged for the audit trail.
func (l *RefundRewardLogic) RefundReward(in *reward.RefundRewardReq) (*reward.Empty, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx,
			"UPDATE `reward_record` SET status='REFUNDED', version=version+1 WHERE id=? AND status NOT IN ('REFUNDED')",
			in.RewardRecordId,
		); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx,
			"UPDATE `outbox` SET status='CANCELLED' WHERE `key`=? AND status='PENDING'",
			fmt.Sprintf("%d", in.RewardRecordId),
		); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx,
			"INSERT INTO `reward_dispatch_log`(reward_record_id, target_service, target_method, request, response, latency_ms, attempt, success, error) VALUES (?,?,?,?,?,?,?,?,?)",
			in.RewardRecordId, "reward", "RefundReward", in.IdempotencyKey, "", 0, 1, 1, in.Reason,
		)
		return err
	})
	if err != nil {
		return nil, err
	}
	bustRewardRecordCache(l.ctx, l.svcCtx, in.RewardRecordId)
	return &reward.Empty{}, nil
}
