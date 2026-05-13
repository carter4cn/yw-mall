package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"mall-activity-rpc/activity"
	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivitySeckillBuyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivitySeckillBuyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivitySeckillBuyLogic {
	return &ActivitySeckillBuyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivitySeckillBuyLogic) ActivitySeckillBuy(req *types.ActivitySeckillBuyReq) (*types.ActivitySeckillBuyResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	payload, _ := json.Marshal(map[string]any{
		"sku_id":   req.SkuId,
		"quantity": qty,
	})

	// per-user-per-sku idempotency: collapses double-tap retries into one row
	clientReqId := fmt.Sprintf("seckill:%d:%d:%d", userId, req.Id, req.SkuId)

	res, err := l.svcCtx.ActivityRpc.Participate(l.ctx, &activity.ParticipateReq{
		ActivityId:      req.Id,
		UserId:          userId,
		Token:           req.Token,
		PayloadJson:     string(payload),
		ClientRequestId: clientReqId,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.ActivitySeckillBuyResp{
		ParticipationId: res.ParticipationId,
		Status:          res.Status,
	}
	if res.DetailJson != "" {
		var d struct {
			OrderNo string `json:"order_no"`
		}
		_ = json.Unmarshal([]byte(res.DetailJson), &d)
		resp.OrderNo = d.OrderNo
	}
	return resp, nil
}
