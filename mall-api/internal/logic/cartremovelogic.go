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

type CartRemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCartRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartRemoveLogic {
	return &CartRemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartRemoveLogic) CartRemove(req *types.CartRemoveReq) error {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	_, err := l.svcCtx.CartRpc.RemoveItem(l.ctx, &cart.RemoveItemReq{
		UserId:    userId,
		ProductId: req.ProductId,
	})
	return err
}
