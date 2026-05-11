package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetProductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetProductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetProductStockLogic {
	return &SetProductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetProductStockLogic) SetProductStock(in *product.SetProductStockReq) (*product.OkResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("product id required")
	}
	if in.Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE product SET stock=? WHERE id=?", in.Stock, in.Id); err != nil {
		return nil, err
	}
	return &product.OkResp{Ok: true}, nil
}
