package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductDetailLogic {
	return &GetProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type productDetailRow struct {
	Id           uint64         `db:"id"`
	Name         string         `db:"name"`
	Description  sql.NullString `db:"description"`
	Price        int64          `db:"price"`
	Stock        int64          `db:"stock"`
	CategoryId   uint64         `db:"category_id"`
	Images       string         `db:"images"`
	ShopId       uint64         `db:"shop_id"`
	Status       int64          `db:"status"`
	CreateTime   time.Time      `db:"create_time"`
	ReviewStatus int64          `db:"review_status"`
	ReviewRemark string         `db:"review_remark"`
	Detail       sql.NullString `db:"detail"`
	Brand        string         `db:"brand"`
	Weight       float64        `db:"weight"`
}

func (l *GetProductDetailLogic) GetProductDetail(in *product.GetProductReq) (*product.ProductDetailResp, error) {
	var r productDetailRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &r,
		`SELECT id, name, description, price, stock, category_id, images, shop_id, status, create_time, review_status, review_remark, detail, brand, weight FROM product WHERE id=? LIMIT 1`, in.Id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	desc := ""
	if r.Description.Valid {
		desc = r.Description.String
	}
	detail := ""
	if r.Detail.Valid {
		detail = r.Detail.String
	}
	return &product.ProductDetailResp{
		Id:           int64(r.Id),
		Name:         r.Name,
		Description:  desc,
		Price:        r.Price,
		Stock:        r.Stock,
		CategoryId:   int64(r.CategoryId),
		Images:       r.Images,
		Status:       int32(r.Status),
		CreateTime:   r.CreateTime.Unix(),
		ShopId:       int64(r.ShopId),
		ReviewStatus: int32(r.ReviewStatus),
		ReviewRemark: r.ReviewRemark,
		Detail:       detail,
		Brand:        r.Brand,
		Weight:       r.Weight,
	}, nil
}
