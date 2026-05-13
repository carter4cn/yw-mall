package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"
	"mall-payment-rpc/paymentclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantInspectReturnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMerchantInspectReturnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantInspectReturnLogic {
	return &MerchantInspectReturnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MerchantInspectReturn records the merchant's inspection result for a returned
// parcel. Passing a type=2 (return_refund) request triggers the actual fund
// callback to mall-payment-rpc and finalizes status=4. Failing (passed=false)
// rejects the refund (status=2). type=3 (exchange) keeps status=1 and waits
// for MerchantShipExchange to advance the state.
func (l *MerchantInspectReturnLogic) MerchantInspectReturn(in *order.MerchantInspectReturnReq) (*order.OkResp, error) {
	r, err := loadRefundById(l.ctx, l.svcCtx, in.RefundId)
	if err != nil {
		return nil, err
	}
	if r.ShopId != in.ShopId {
		return nil, errors.New("shop mismatch")
	}
	if r.Status != 1 {
		return nil, errors.New("refund not in approved state")
	}
	if r.ReturnTrackingNo == "" {
		return nil, errors.New("user has not yet shipped return")
	}
	now := time.Now().Unix()
	passedFlag := int64(1)
	if !in.Passed {
		passedFlag = 2
	}
	if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE refund_request SET return_received_time = ?, return_inspection_passed = ?, merchant_remark = ?, update_time = ? WHERE id = ?",
		now, passedFlag, in.Remark, now, in.RefundId,
	); err != nil {
		return nil, err
	}

	if !in.Passed {
		// Rejected: refund flips to status=2 (rejected).
		if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 2, update_time = ? WHERE id = ?",
			now, in.RefundId,
		); err != nil {
			return nil, err
		}
		return &order.OkResp{Ok: true}, nil
	}

	// Passed + type=2 return_refund: trigger payment callback.
	if r.RefundType == 2 {
		if l.svcCtx.PaymentRpc == nil {
			return nil, errors.New("payment rpc not configured")
		}
		refundNo := r.RefundNo
		if refundNo == "" {
			refundNo = generateRefundNo()
		}
		if _, err := l.svcCtx.PaymentRpc.ExecuteRefund(l.ctx, &paymentclient.ExecuteRefundReq{
			OrderId:  r.OrderId,
			OrderNo:  r.OrderNo,
			ShopId:   r.ShopId,
			Amount:   r.Amount,
			Reason:   fmt.Sprintf("return-refund: %s", r.Reason),
			RefundNo: refundNo,
		}); err != nil {
			return nil, err
		}
		if _, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE refund_request SET status = 4, refund_no = ?, refund_complete_time = ?, update_time = ? WHERE id = ?",
			refundNo, now, now, in.RefundId,
		); err != nil {
			return nil, err
		}
	}
	// type=3 exchange: stay status=1, wait for MerchantShipExchange.
	return &order.OkResp{Ok: true}, nil
}
