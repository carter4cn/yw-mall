package logic

import (
	"context"
	"errors"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantRejectRefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMerchantRejectRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantRejectRefundLogic {
	return &MerchantRejectRefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MerchantRejectRefundLogic) MerchantRejectRefund(in *order.MerchantRejectRefundReq) (*order.OkResp, error) {
	res, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE `order` SET refund_status=3, refund_reason=? WHERE id=? AND shop_id=? AND refund_status=1",
		in.Reason, in.Id, in.ShopId)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, errors.New("no pending refund for this order")
	}
	return &order.OkResp{Ok: true}, nil
}
