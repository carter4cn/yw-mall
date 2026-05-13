package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppealRefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppealRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppealRefundLogic {
	return &AppealRefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppealRefundLogic) AppealRefund(req *types.AppealRefundReq) (*types.OkResp, error) {
	uid := uidFromContext(l.ctx)
	if uid == 0 {
		return nil, errors.New("unauthorized")
	}
	if _, err := l.svcCtx.OrderRpc.UserAppealRefund(l.ctx, &order.UserAppealRefundReq{
		RefundId: req.Id,
		UserId:   uid,
		Reason:   req.Reason,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
