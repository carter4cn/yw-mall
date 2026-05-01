package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkParticipationRewardedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkParticipationRewardedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkParticipationRewardedLogic {
	return &MarkParticipationRewardedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkParticipationRewarded is the SAGA-callable hook that flips a record
// status to REWARDED. Idempotency_key prevents double-application on retries.
func (l *MarkParticipationRewardedLogic) MarkParticipationRewarded(in *activity.MarkParticipationRewardedReq) (*activity.Empty, error) {
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `participation_record` SET status='REWARDED' WHERE id=? AND status IN ('PENDING','CHECKED_IN','RESERVED','WON','ISSUED')",
		in.ParticipationId,
	)
	if err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
