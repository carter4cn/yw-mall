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

type ActivityLotterySpinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityLotterySpinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityLotterySpinLogic {
	return &ActivityLotterySpinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityLotterySpinLogic) ActivityLotterySpin(req *types.ActivityLotterySpinReq) (*types.ActivityLotterySpinResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	res, err := l.svcCtx.ActivityRpc.Participate(l.ctx, &activity.ParticipateReq{
		ActivityId: req.Id,
		UserId:     userId,
		Token:      req.Token,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.ActivityLotterySpinResp{
		ParticipationId: res.ParticipationId,
		Status:          res.Status,
	}
	if res.DetailJson != "" {
		var d struct {
			PrizeId   int64  `json:"prize_id"`
			PrizeCode string `json:"prize_code"`
			PrizeName string `json:"prize_name"`
		}
		_ = json.Unmarshal([]byte(res.DetailJson), &d)
		resp.PrizeCode = d.PrizeCode
		resp.PrizeName = d.PrizeName
	}
	return resp, nil
}
