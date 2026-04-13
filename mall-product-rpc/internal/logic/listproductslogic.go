package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type productRow struct {
	Id          uint64         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Price       int64          `db:"price"`
	Stock       int64          `db:"stock"`
	CategoryId  uint64         `db:"category_id"`
	Images      string         `db:"images"`
	Status      int64          `db:"status"`
	CreateTime  time.Time      `db:"create_time"`
}

func (l *ListProductsLogic) ListProducts(in *product.ListProductsReq) (*product.ListProductsResp, error) {
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

	var (
		rows  []productRow
		total int64
	)

	if in.CategoryId > 0 {
		countQuery := "SELECT COUNT(*) FROM product WHERE category_id = ?"
		if err := conn.QueryRowCtx(l.ctx, &total, countQuery, in.CategoryId); err != nil {
			return nil, err
		}

		query := fmt.Sprintf("SELECT id, name, description, price, stock, category_id, images, status, create_time FROM product WHERE category_id = ? ORDER BY id DESC LIMIT %d OFFSET %d", pageSize, offset)
		if err := conn.QueryRowsCtx(l.ctx, &rows, query, in.CategoryId); err != nil {
			return nil, err
		}
	} else {
		countQuery := "SELECT COUNT(*) FROM product"
		if err := conn.QueryRowCtx(l.ctx, &total, countQuery); err != nil {
			return nil, err
		}

		query := fmt.Sprintf("SELECT id, name, description, price, stock, category_id, images, status, create_time FROM product ORDER BY id DESC LIMIT %d OFFSET %d", pageSize, offset)
		if err := conn.QueryRowsCtx(l.ctx, &rows, query); err != nil {
			return nil, err
		}
	}

	products := make([]*product.GetProductResp, 0, len(rows))
	for _, r := range rows {
		description := ""
		if r.Description.Valid {
			description = r.Description.String
		}
		products = append(products, &product.GetProductResp{
			Id:          int64(r.Id),
			Name:        r.Name,
			Description: description,
			Price:       r.Price,
			Stock:       r.Stock,
			CategoryId:  int64(r.CategoryId),
			Images:      r.Images,
			Status:      int32(r.Status),
			CreateTime:  r.CreateTime.Unix(),
		})
	}

	return &product.ListProductsResp{
		Products: products,
		Total:    total,
	}, nil
}
