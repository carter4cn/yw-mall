package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopRefundRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopRefundRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopRefundRequestsLogic {
	return &ListShopRefundRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopRefundRequestsLogic) ListShopRefundRequests(in *order.ListShopRefundRequestsReq) (*order.ListRefundRequestsResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	where := "shop_id = ?"
	args := []any{in.ShopId}
	if in.Status >= 0 {
		where += " AND status = ?"
		args = append(args, in.Status)
	}

	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM refund_request WHERE "+where, args...); err != nil {
		return nil, err
	}

	var rows []*refundRow
	args = append(args, pageSize, (page-1)*pageSize)
	q := "SELECT id, order_id, order_no, user_id, shop_id, amount, reason, evidence, items, status, merchant_user_id, merchant_remark, merchant_handle_time, admin_id, admin_remark, admin_handle_time, appeal_reason, appeal_time, refund_no, refund_complete_time, create_time FROM refund_request WHERE " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}

	out := make([]*order.RefundRequest, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRefundProto(r))
	}
	return &order.ListRefundRequestsResp{Requests: out, Total: total}, nil
}
