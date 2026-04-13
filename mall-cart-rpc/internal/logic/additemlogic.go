package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddItemLogic {
	return &AddItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddItemLogic) AddItem(in *cart.AddItemReq) (*cart.AddItemResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO cart_item (user_id, product_id, quantity, selected) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE quantity = quantity + ?",
		in.UserId, in.ProductId, in.Quantity, in.Quantity,
	)
	if err != nil {
		return nil, err
	}

	return &cart.AddItemResp{}, nil
}
