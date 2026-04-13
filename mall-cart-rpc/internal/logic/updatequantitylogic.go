package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateQuantityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateQuantityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateQuantityLogic {
	return &UpdateQuantityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateQuantityLogic) UpdateQuantity(in *cart.UpdateQuantityReq) (*cart.UpdateQuantityResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE cart_item SET quantity = ? WHERE user_id = ? AND product_id = ?",
		in.Quantity, in.UserId, in.ProductId,
	)
	if err != nil {
		return nil, err
	}

	return &cart.UpdateQuantityResp{}, nil
}
