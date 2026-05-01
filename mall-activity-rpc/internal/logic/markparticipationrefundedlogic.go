package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkParticipationRefundedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkParticipationRefundedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkParticipationRefundedLogic {
	return &MarkParticipationRefundedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkParticipationRefundedLogic) MarkParticipationRefunded(in *activity.MarkParticipationRefundedReq) (*activity.Empty, error) {
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `participation_record` SET status='PENDING_REFUND' WHERE id=?",
		in.ParticipationId,
	)
	if err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
