package logic

import (
	"context"

	"mall-common/errorx"
	"mall-shop-rpc/internal/model"
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
	s, err := l.svcCtx.ShopModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.ShopNotFound)
		}
		return nil, err
	}
	return &shop.GetShopResp{Shop: toShopProto(s)}, nil
}
