package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListReviewProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListReviewProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListReviewProductsLogic {
	return &AdminListReviewProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminListReviewProductsLogic) AdminListReviewProducts(in *product.AdminListReviewProductsReq) (*product.ListProductsResp, error) {
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM product WHERE review_status=?", in.ReviewStatus); err != nil {
		return nil, err
	}

	var rows []productRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, name, description, price, stock, category_id, images, shop_id, status, create_time FROM product WHERE review_status=? ORDER BY id DESC LIMIT ? OFFSET ?",
		in.ReviewStatus, pageSize, offset); err != nil {
		return nil, err
	}

	out := make([]*product.GetProductResp, 0, len(rows))
	for _, r := range rows {
		out = append(out, toProductProto(r))
	}
	return &product.ListProductsResp{Products: out, Total: total}, nil
}
