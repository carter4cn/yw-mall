// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"github.com/zeromicro/go-zero/core/logx"
	"mall-api/internal/middleware"
	"mall-api/internal/svc"
)

type CartClearLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCartClearLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartClearLogic {
	return &CartClearLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartClearLogic) CartClear() error {
	userId := middleware.UidFromCtx(l.ctx)

	_, err := l.svcCtx.CartRpc.ClearCart(l.ctx, &cart.ClearCartReq{
		UserId: userId,
	})
	return err
}
