package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMerchantListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantListProductsLogic {
	return &MerchantListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MerchantListProductsLogic) MerchantListProducts(in *product.MerchantListProductsReq) (*product.ListProductsResp, error) {
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "shop_id=?"
	args := []any{in.ShopId}
	if in.Status >= 0 {
		where += " AND status=?"
		args = append(args, in.Status)
	}
	if in.ReviewStatus >= 0 {
		where += " AND review_status=?"
		args = append(args, in.ReviewStatus)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM product WHERE "+where, args...); err != nil {
		return nil, err
	}

	listArgs := append(args, pageSize, offset)
	var rows []productRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, name, description, price, stock, category_id, images, shop_id, status, create_time FROM product WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?",
		listArgs...); err != nil {
		return nil, err
	}

	out := make([]*product.GetProductResp, 0, len(rows))
	for _, r := range rows {
		out = append(out, toProductProto(r))
	}
	return &product.ListProductsResp{Products: out, Total: total}, nil
}
