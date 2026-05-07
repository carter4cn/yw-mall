package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopLogic {
	return &GetShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetShopLogic) GetShop(in *shop.GetShopReq) (*shop.GetShopResp, error) {
	// todo: add your logic here and delete this line

	return &shop.GetShopResp{}, nil
}
