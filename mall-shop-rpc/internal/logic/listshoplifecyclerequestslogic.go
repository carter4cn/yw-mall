package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopLifecycleRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopLifecycleRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopLifecycleRequestsLogic {
	return &ListShopLifecycleRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopLifecycleRequestsLogic) ListShopLifecycleRequests(in *shop.ListShopLifecycleRequestsReq) (*shop.ListShopLifecycleRequestsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size

	where := "1=1"
	args := []any{}
	if in.Status >= 0 {
		where += " AND status=?"
		args = append(args, in.Status)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM shop_lifecycle_request WHERE "+where, args...); err != nil {
		return nil, err
	}

	listArgs := append(args, size, offset)
	var rows []*lifecycleRequestRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+lifecycleRequestCols+" FROM shop_lifecycle_request WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", listArgs...); err != nil {
		return nil, err
	}

	out := make([]*shop.ShopLifecycleRequest, 0, len(rows))
	for _, r := range rows {
		out = append(out, toLifecycleRequestProto(r))
	}
	return &shop.ListShopLifecycleRequestsResp{Requests: out, Total: total}, nil
}
