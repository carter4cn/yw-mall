package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantShipExchangeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMerchantShipExchangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantShipExchangeLogic {
	return &MerchantShipExchangeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MerchantShipExchange spawns a free replacement order for an exchange (type=3)
// refund whose inspection has passed. The new order is created in status=2
// (shipped) with total_amount=0 and the merchant's tracking number stamped.
// The refund itself transitions to status=4 (completed) with no fund movement.
func (l *MerchantShipExchangeLogic) MerchantShipExchange(in *order.MerchantShipExchangeReq) (*order.MerchantShipExchangeResp, error) {
	r, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if r.ShopId != in.ShopId {
		return nil, errors.New("shop mismatch")
	}
	if r.RefundType != 3 {
		return nil, errors.New("not an exchange request")
	}
	if r.ReturnInspectionPassed != 1 {
		return nil, errors.New("inspection has not passed")
	}
	if r.ExchangeNewOrderId != 0 {
		return nil, errors.New("exchange order already shipped")
	}

	now := time.Now().Unix()
	newOrderNo := fmt.Sprintf("EX%d", now)

	// Clone original order's user/address/receiver fields. status=2 (shipped),
	// total_amount=0, settle_status=2 (skipped) so settlement never touches it.
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO `order` (order_no, user_id, total_amount, status, shop_id, tracking_no, carrier, address_id, receiver_name, receiver_phone, receiver_province, receiver_city, receiver_district, receiver_detail, pay_time, ship_time, settle_status) "+
			"SELECT ?, user_id, 0, 2, shop_id, ?, ?, address_id, receiver_name, receiver_phone, receiver_province, receiver_city, receiver_district, receiver_detail, ?, ?, 2 FROM `order` WHERE id = ?",
		newOrderNo, in.TrackingNo, in.Carrier, now, now, r.OrderId,
	)
	if err != nil {
		return nil, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if newId == 0 {
		return nil, errors.New("failed to create exchange order")
	}

	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET exchange_new_order_id = ?, status = 4, refund_complete_time = ?, update_time = ? WHERE id = ?",
		newId, now, now, in.RefundId,
	); err != nil {
		return nil, err
	}

	return &order.MerchantShipExchangeResp{
		NewOrderId: newId,
		NewOrderNo: newOrderNo,
	}, nil
}
