package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SearchProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchProductsLogic {
	return &SearchProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchProductsLogic) SearchProducts(in *product.SearchProductsReq) (*product.SearchProductsResp, error) {
	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)

	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	keyword := "%" + in.Keyword + "%"

	var total int64
	countQuery := "SELECT COUNT(*) FROM product WHERE name LIKE ?"
	if err := conn.QueryRowCtx(l.ctx, &total, countQuery, keyword); err != nil {
		return nil, err
	}

	var rows []productRow
	query := fmt.Sprintf("SELECT id, name, description, price, stock, category_id, images, shop_id, status, create_time FROM product WHERE name LIKE ? ORDER BY id DESC LIMIT %d OFFSET %d", pageSize, offset)
	if err := conn.QueryRowsCtx(l.ctx, &rows, query, keyword); err != nil {
		return nil, err
	}

	products := make([]*product.GetProductResp, 0, len(rows))
	for _, r := range rows {
		products = append(products, toProductProto(r))
	}

	return &product.SearchProductsResp{
		Products: products,
		Total:    total,
	}, nil
}
