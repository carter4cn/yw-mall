package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateShopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShopLogic {
	return &UpdateShopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateShopLogic) UpdateShop(in *shop.UpdateShopReq) (*shop.OkResp, error) {
	// todo: add your logic here and delete this line

	return &shop.OkResp{}, nil
}
