package logic

import (
	"context"
	"errors"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserShipReturnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserShipReturnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserShipReturnLogic {
	return &UserShipReturnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserShipReturn lets the buyer record their return-shipment tracking number
// after the merchant has approved a return_refund (type=2) or exchange (type=3)
// refund request (status=1 approved). Refund_only (type=1) never reaches here.
func (l *UserShipReturnLogic) UserShipReturn(in *order.UserShipReturnReq) (*order.OkResp, error) {
	if in.RefundId == 0 || in.UserId == 0 || in.TrackingNo == "" {
		return nil, errors.New("refund_id, user_id, tracking_no required")
	}
	r, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if r.UserId != in.UserId {
		return nil, errors.New("refund does not belong to user")
	}
	if r.Status != 1 {
		return nil, errors.New("refund not in approved state")
	}
	if r.RefundType != 2 && r.RefundType != 3 {
		return nil, errors.New("refund type does not require return shipment")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET return_tracking_no = ?, return_carrier = ?, return_ship_time = ?, update_time = ? WHERE id = ? AND status = 1",
		in.TrackingNo, in.Carrier, now, now, in.RefundId,
	); err != nil {
		return nil, err
	}
	return &order.OkResp{Ok: true}, nil
}
