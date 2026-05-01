package logic

import (
	"context"
	"fmt"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmLogic {
	return &ConfirmLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Confirm flips PENDING/DISPATCHED → CONFIRMED. The status guard in the WHERE
// clause makes this idempotent: re-confirming a CONFIRMED row is a no-op.
// Refusing to confirm a FAILED/REFUNDED row protects the SAGA invariant.
func (l *ConfirmLogic) Confirm(in *reward.ConfirmReq) (*reward.Empty, error) {
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `reward_record` SET status='CONFIRMED', version=version+1 WHERE id=? AND status IN ('PENDING','DISPATCHED')",
		in.RewardRecordId,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var status string
		_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &status,
			"SELECT status FROM `reward_record` WHERE id=? LIMIT 1 FOR UPDATE", in.RewardRecordId)
		if status == "FAILED" || status == "REFUNDED" {
			return nil, fmt.Errorf("reward %d is in terminal state %s; cannot confirm", in.RewardRecordId, status)
		}
	}
	bustRewardRecordCache(l.ctx, l.svcCtx, in.RewardRecordId)
	return &reward.Empty{}, nil
}
