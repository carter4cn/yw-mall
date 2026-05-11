package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyApplicationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyApplicationsLogic {
	return &ListMyApplicationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMyApplicationsLogic) ListMyApplications(in *shop.ListMyApplicationsReq) (*shop.ListShopApplicationsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM shop_application WHERE user_id=?", in.UserId); err != nil {
		return nil, err
	}

	var rows []*applicationRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+applicationCols+" FROM shop_application WHERE user_id=? ORDER BY id DESC LIMIT ? OFFSET ?",
		in.UserId, size, offset); err != nil {
		return nil, err
	}

	out := make([]*shop.ShopApplication, 0, len(rows))
	for _, r := range rows {
		out = append(out, toApplicationProto(r))
	}
	return &shop.ListShopApplicationsResp{Applications: out, Total: total}, nil
}
