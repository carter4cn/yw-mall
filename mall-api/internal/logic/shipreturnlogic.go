package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShipReturnLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShipReturnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipReturnLogic {
	return &ShipReturnLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShipReturn proxies UserShipReturn to mall-order-rpc with JWT-derived user_id.
func (l *ShipReturnLogic) ShipReturn(req *types.ShipReturnReq) (*types.OkResp, error) {
	uid := uidFromContext(l.ctx)
	if uid == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.TrackingNo == "" {
		return nil, errors.New("tracking_no required")
	}
	if _, err := l.svcCtx.OrderRpc.UserShipReturn(l.ctx, &order.UserShipReturnReq{
		RefundId:   req.Id,
		UserId:     uid,
		TrackingNo: req.TrackingNo,
		Carrier:    req.Carrier,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
