package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetProductStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetProductStatusLogic {
	return &SetProductStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetProductStatusLogic) SetProductStatus(in *product.SetProductStatusReq) (*product.OkResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("product id required")
	}
	var ownerShopId int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &ownerShopId, "SELECT shop_id FROM product WHERE id=? LIMIT 1", in.Id); err != nil {
		return nil, err
	}
	if in.ShopId > 0 && ownerShopId != in.ShopId {
		return nil, errors.New("product not owned by shop")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE product SET status=? WHERE id=?", in.Status, in.Id); err != nil {
		return nil, err
	}
	return &product.OkResp{Ok: true}, nil
}
