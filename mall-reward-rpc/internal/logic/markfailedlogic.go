package logic

import (
	"context"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkFailedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkFailedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkFailedLogic {
	return &MarkFailedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkFailed transitions a non-confirmed reward to FAILED and writes a dispatch_log
// entry capturing the reason — that log row is what an operator looks at after a
// DLQ alert fires.
func (l *MarkFailedLogic) MarkFailed(in *reward.MarkFailedReq) (*reward.Empty, error) {
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `reward_record` SET status='FAILED', version=version+1 WHERE id=? AND status NOT IN ('CONFIRMED','REFUNDED')",
		in.RewardRecordId,
	); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO `reward_dispatch_log`(reward_record_id, target_service, target_method, request, response, latency_ms, attempt, success, error) VALUES (?,?,?,?,?,?,?,?,?)",
		in.RewardRecordId, "reward", "MarkFailed", "", "", 0, 1, 0, in.Reason,
	); err != nil {
		return nil, err
	}
	bustRewardRecordCache(l.ctx, l.svcCtx, in.RewardRecordId)
	return &reward.Empty{}, nil
}
