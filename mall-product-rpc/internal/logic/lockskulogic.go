package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type LockSkuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockSkuLogic {
	return &LockSkuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LockSkuLogic) LockSku(in *product.LockSkuReq) (*product.LockSkuResp, error) {
	// todo: add your logic here and delete this line

	return &product.LockSkuResp{}, nil
}
