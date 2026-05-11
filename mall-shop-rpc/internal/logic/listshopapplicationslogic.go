package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopApplicationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopApplicationsLogic {
	return &ListShopApplicationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopApplicationsLogic) ListShopApplications(in *shop.ListShopApplicationsReq) (*shop.ListShopApplicationsResp, error) {
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
		"SELECT COUNT(*) FROM shop_application WHERE "+where, args...); err != nil {
		return nil, err
	}

	listArgs := append(args, size, offset)
	var rows []*applicationRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+applicationCols+" FROM shop_application WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", listArgs...); err != nil {
		return nil, err
	}

	out := make([]*shop.ShopApplication, 0, len(rows))
	for _, r := range rows {
		out = append(out, toApplicationProto(r))
	}
	return &shop.ListShopApplicationsResp{Applications: out, Total: total}, nil
}
