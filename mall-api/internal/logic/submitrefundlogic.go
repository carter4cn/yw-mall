package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitRefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitRefundLogic {
	return &SubmitRefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitRefundLogic) SubmitRefund(req *types.SubmitRefundReq) (*types.SubmitRefundResp, error) {
	uid := uidFromContext(l.ctx)
	if uid == 0 {
		return nil, errors.New("unauthorized")
	}
	items := make([]*order.RefundItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &order.RefundItem{
			SkuId:    it.SkuId,
			SkuName:  it.SkuName,
			Quantity: it.Quantity,
			Amount:   it.Amount,
		})
	}
	resp, err := l.svcCtx.OrderRpc.SubmitRefundRequest(l.ctx, &order.SubmitRefundRequestReq{
		OrderId:  req.OrderId,
		UserId:   uid,
		Amount:   req.Amount,
		Reason:   req.Reason,
		Evidence: req.Evidence,
		Items:    items,
	})
	if err != nil {
		return nil, err
	}
	return &types.SubmitRefundResp{RefundId: resp.RefundId}, nil
}
