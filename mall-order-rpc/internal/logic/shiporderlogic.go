package logic

import (
	"context"
	"errors"
	"strings"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShipOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ShipOrderLogic) ShipOrder(in *order.ShipOrderReq) (*order.OkResp, error) {
	if strings.TrimSpace(in.TrackingNo) == "" || strings.TrimSpace(in.Carrier) == "" {
		return nil, errors.New("carrier and tracking_no are required")
	}
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE `order` SET status=2, tracking_no=?, carrier=? WHERE id=? AND shop_id=? AND status=1",
		in.TrackingNo, in.Carrier, in.Id, in.ShopId)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, errors.New("order not in shippable state or not owned by shop")
	}
	return &order.OkResp{Ok: true}, nil
}
