package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowedShopsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowedShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowedShopsLogic {
	return &ListFollowedShopsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowedShopsLogic) ListFollowedShops(in *shop.ListFollowedShopsReq) (*shop.ListShopsResp, error) {
	// todo: add your logic here and delete this line

	return &shop.ListShopsResp{}, nil
}
