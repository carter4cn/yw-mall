package logic

import (
	"context"
	"encoding/json"

	"mall-activity-rpc/activity"
	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivitySignInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivitySignInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivitySignInLogic {
	return &ActivitySignInLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivitySignInLogic) ActivitySignIn(req *types.ActivitySignInReq) (*types.ActivitySignInResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	res, err := l.svcCtx.ActivityRpc.Participate(l.ctx, &activity.ParticipateReq{
		ActivityId: req.Id,
		UserId:     userId,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.ActivitySignInResp{Status: res.Status}
	if res.DetailJson != "" {
		var d struct {
			PointsAwarded int64 `json:"points_awarded"`
			StreakDays    int32 `json:"streak_days"`
		}
		_ = json.Unmarshal([]byte(res.DetailJson), &d)
		resp.PointsAwarded = d.PointsAwarded
		resp.StreakDays = d.StreakDays
	}
	return resp, nil
}
