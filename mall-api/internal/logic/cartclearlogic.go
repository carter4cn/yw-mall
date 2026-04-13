// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"

	"mall-cart-rpc/cart"
	"github.com/zeromicro/go-zero/core/logx"
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
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	_, err := l.svcCtx.CartRpc.ClearCart(l.ctx, &cart.ClearCartReq{
		UserId: userId,
	})
	return err
}
