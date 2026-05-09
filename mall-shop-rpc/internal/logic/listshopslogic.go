package logic

import (
	"context"

	"mall-shop-rpc/internal/model"
	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopsLogic {
	return &ListShopsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopsLogic) ListShops(in *shop.ListShopsReq) (*shop.ListShopsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size
	rows, err := l.queryList(offset, size)
	if err != nil {
		return nil, err
	}
	total, _ := l.countAll()
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: total}, nil
}

func (l *ListShopsLogic) queryList(offset, size int32) ([]*model.Shop, error) {
	var rows []*model.Shop
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, "SELECT * FROM shop WHERE status=1 ORDER BY id DESC LIMIT ?, ?", offset, size)
	return rows, err
}

func (l *ListShopsLogic) countAll() (int64, error) {
	var n int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &n, "SELECT COUNT(*) FROM shop WHERE status=1")
	return n, err
}
