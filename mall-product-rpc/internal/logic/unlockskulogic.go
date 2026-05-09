package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockSkuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockSkuLogic {
	return &UnlockSkuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnlockSkuLogic) UnlockSku(in *product.UnlockSkuReq) (*product.UnlockSkuResp, error) {
	// todo: add your logic here and delete this line

	return &product.UnlockSkuResp{}, nil
}
