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

type ActivityCouponClaimLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityCouponClaimLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityCouponClaimLogic {
	return &ActivityCouponClaimLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityCouponClaimLogic) ActivityCouponClaim(req *types.ActivityCouponClaimReq) (*types.ActivityCouponClaimResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	var payload string
	if req.CouponCodeId != 0 {
		b, _ := json.Marshal(map[string]any{"coupon_code_id": req.CouponCodeId})
		payload = string(b)
	}

	res, err := l.svcCtx.ActivityRpc.Participate(l.ctx, &activity.ParticipateReq{
		ActivityId:  req.Id,
		UserId:      userId,
		PayloadJson: payload,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.ActivityCouponClaimResp{
		ParticipationId: res.ParticipationId,
		Status:          res.Status,
	}
	if res.DetailJson != "" {
		var d struct {
			CouponCode string `json:"coupon_code"`
		}
		_ = json.Unmarshal([]byte(res.DetailJson), &d)
		resp.CouponCode = d.CouponCode
	}
	return resp, nil
}
