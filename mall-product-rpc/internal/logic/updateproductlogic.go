package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductLogic) UpdateProduct(in *product.UpdateProductReq) (*product.OkResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("product id required")
	}
	// verify ownership
	var ownerShopId int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &ownerShopId, "SELECT shop_id FROM product WHERE id=? LIMIT 1", in.Id); err != nil {
		return nil, err
	}
	if in.ShopId > 0 && ownerShopId != in.ShopId {
		return nil, errors.New("product not owned by shop")
	}
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`UPDATE product SET name=?, description=?, price=?, category_id=?, images=?, detail=?, brand=?, weight=?, review_status=0, review_remark='' WHERE id=?`,
		in.Name, in.Description, in.Price, in.CategoryId, in.Images, in.Detail, in.Brand, in.Weight, in.Id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("product not found")
	}
	return &product.OkResp{Ok: true}, nil
}
