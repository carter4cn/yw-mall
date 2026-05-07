package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRecommendedShopsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRecommendedShopsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRecommendedShopsLogic {
	return &ListRecommendedShopsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRecommendedShopsLogic) ListRecommendedShops(in *shop.ListRecommendedShopsReq) (*shop.ListShopsResp, error) {
	// todo: add your logic here and delete this line

	return &shop.ListShopsResp{}, nil
}
