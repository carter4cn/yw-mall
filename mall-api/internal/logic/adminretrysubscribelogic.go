// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminRetrySubscribeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRetrySubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRetrySubscribeLogic {
	return &AdminRetrySubscribeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminRetrySubscribeLogic) AdminRetrySubscribe(req *types.AdminRetrySubscribeReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.LogisticsRpc.RetrySubscribe(l.ctx, &logisticspb.RetrySubscribeReq{ShipmentId: req.Id}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
