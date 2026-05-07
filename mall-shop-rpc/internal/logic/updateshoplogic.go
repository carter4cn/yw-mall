package logic

import (
	"context"

	"mall-common/errorx"
	"mall-shop-rpc/internal/model"
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
	s, err := l.svcCtx.ShopModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.ShopNotFound)
		}
		return nil, err
	}
	if in.Name != "" {
		s.Name = in.Name
	}
	if in.Logo != "" {
		s.Logo = in.Logo
	}
	if in.Banner != "" {
		s.Banner = in.Banner
	}
	if in.Description != "" {
		s.Description = in.Description
	}
	if err := l.svcCtx.ShopModel.Update(l.ctx, s); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
