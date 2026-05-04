// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	orderpb "mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminMarkShippedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminMarkShippedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminMarkShippedLogic {
	return &AdminMarkShippedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminMarkShippedLogic) AdminMarkShipped(req *types.AdminMarkShippedReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.OrderRpc.MarkShipped(l.ctx, &orderpb.MarkShippedReq{
		OrderId:    req.Id,
		TrackingNo: req.TrackingNo,
		Carrier:    req.Carrier,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
