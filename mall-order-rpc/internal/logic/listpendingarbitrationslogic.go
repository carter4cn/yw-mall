package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPendingArbitrationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPendingArbitrationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPendingArbitrationsLogic {
	return &ListPendingArbitrationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListPendingArbitrations returns refund requests with status=3 (arbitrating).
func (l *ListPendingArbitrationsLogic) ListPendingArbitrations(in *order.ListPendingArbitrationsReq) (*order.ListRefundRequestsResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM refund_request WHERE status = 3"); err != nil {
		return nil, err
	}

	var rows []*refundRow
	q := "SELECT id, order_id, order_no, user_id, shop_id, amount, reason, evidence, items, status, merchant_user_id, merchant_remark, merchant_handle_time, admin_id, admin_remark, admin_handle_time, appeal_reason, appeal_time, refund_no, refund_complete_time, create_time FROM refund_request WHERE status = 3 ORDER BY appeal_time ASC, id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows, q, pageSize, (page-1)*pageSize); err != nil {
		return nil, err
	}

	out := make([]*order.RefundRequest, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRefundProto(r))
	}
	return &order.ListRefundRequestsResp{Requests: out, Total: total}, nil
}
