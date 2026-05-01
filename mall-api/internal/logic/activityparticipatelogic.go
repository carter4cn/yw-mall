package logic

import (
	"context"
	"encoding/json"

	"mall-activity-rpc/activity"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivityParticipateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityParticipateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityParticipateLogic {
	return &ActivityParticipateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityParticipateLogic) ActivityParticipate(req *types.ActivityParticipateReq) (*types.ActivityParticipateResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.ActivityRpc.Participate(l.ctx, &activity.ParticipateReq{
		ActivityId:      req.Id,
		UserId:          userId,
		Token:           req.Token,
		PayloadJson:     req.PayloadJson,
		ClientRequestId: req.ClientRequestId,
	})
	if err != nil {
		return nil, err
	}
	return &types.ActivityParticipateResp{
		ParticipationId:    res.ParticipationId,
		WorkflowInstanceId: res.WorkflowInstanceId,
		Status:             res.Status,
		DetailJson:         res.DetailJson,
	}, nil
}
