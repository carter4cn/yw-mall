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

type CartUpdateQuantityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCartUpdateQuantityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartUpdateQuantityLogic {
	return &CartUpdateQuantityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartUpdateQuantityLogic) CartUpdateQuantity(req *types.CartUpdateQuantityReq) error {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	_, err := l.svcCtx.CartRpc.UpdateQuantity(l.ctx, &cart.UpdateQuantityReq{
		UserId:    userId,
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
	})
	return err
}
