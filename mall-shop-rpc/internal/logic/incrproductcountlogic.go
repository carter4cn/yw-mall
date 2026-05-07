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
	if in.ShopId == 0 || in.Delta == 0 {
		return &shop.OkResp{Ok: true}, nil
	}
	_, err := l.svcCtx.DB.ExecCtx(l.ctx, "UPDATE shop SET product_count = product_count + ? WHERE id = ?", in.Delta, in.ShopId)
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
