package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShopLogic {
	return &CreateShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateShopLogic) CreateShop(in *shop.CreateShopReq) (*shop.CreateShopResp, error) {
	// todo: add your logic here and delete this line

	return &shop.CreateShopResp{}, nil
}
