// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-cart-rpc/cart"

	"github.com/zeromicro/go-zero/core/logx"
)

type CartAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCartAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartAddLogic {
	return &CartAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartAddLogic) CartAdd(req *types.CartAddReq) error {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	_, err := l.svcCtx.CartRpc.AddItem(l.ctx, &cart.AddItemReq{
		UserId:    userId,
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
	})
	return err
}
