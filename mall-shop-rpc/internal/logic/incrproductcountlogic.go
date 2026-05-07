package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrProductCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrProductCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrProductCountLogic {
	return &IncrProductCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncrProductCountLogic) IncrProductCount(in *shop.IncrProductCountReq) (*shop.OkResp, error) {
	// todo: add your logic here and delete this line

	return &shop.OkResp{}, nil
}
