package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListShopProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopProductsLogic {
	return &ListShopProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopProductsLogic) ListShopProducts(in *product.ListShopProductsReq) (*product.ListProductsResp, error) {
	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := conn.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM product WHERE shop_id=? AND status=1", in.ShopId); err != nil {
		return nil, err
	}

	var rows []productRow
	q := fmt.Sprintf("SELECT id, name, description, price, stock, category_id, images, shop_id, status, create_time FROM product WHERE shop_id=? AND status=1 ORDER BY id DESC LIMIT %d OFFSET %d", pageSize, offset)
	if err := conn.QueryRowsCtx(l.ctx, &rows, q, in.ShopId); err != nil {
		return nil, err
	}

	products := make([]*product.GetProductResp, 0, len(rows))
	for _, r := range rows {
		products = append(products, toProductProto(r))
	}
	return &product.ListProductsResp{Products: products, Total: total}, nil
}
