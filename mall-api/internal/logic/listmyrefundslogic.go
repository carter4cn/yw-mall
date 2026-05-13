package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyRefundsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyRefundsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRefundsLogic {
	return &ListMyRefundsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyRefundsLogic) ListMyRefunds(req *types.ListMyRefundsReq) (*types.ListRefundsResp, error) {
	uid := uidFromContext(l.ctx)
	if uid == 0 {
		return nil, errors.New("unauthorized")
	}
	resp, err := l.svcCtx.OrderRpc.ListUserRefundRequests(l.ctx, &order.ListUserRefundRequestsReq{
		UserId:   uid,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.RefundRequestDTO, 0, len(resp.Requests))
	for _, r := range resp.Requests {
		out = append(out, refundProtoToDTO(r))
	}
	return &types.ListRefundsResp{Requests: out, Total: resp.Total}, nil
}
