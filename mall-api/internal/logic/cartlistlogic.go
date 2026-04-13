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

type CartListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCartListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartListLogic {
	return &CartListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartListLogic) CartList() (resp *types.CartListResp, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.CartRpc.ListItems(l.ctx, &cart.ListItemsReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.CartItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, types.CartItem{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
			Selected:  item.Selected,
		})
	}

	return &types.CartListResp{
		Items: items,
	}, nil
}
