package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRefundLogic {
	return &GetRefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRefundLogic) GetRefund(req *types.GetRefundReq) (*types.RefundRequestDTO, error) {
	uid := uidFromContext(l.ctx)
	if uid == 0 {
		return nil, errors.New("unauthorized")
	}
	r, err := l.svcCtx.OrderRpc.GetRefundRequest(l.ctx, &order.GetRefundRequestReq{Id: req.Id, UserId: uid})
	if err != nil {
		return nil, err
	}
	dto := refundProtoToDTO(r)
	return &dto, nil
}
