package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveItemLogic {
	return &RemoveItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveItemLogic) RemoveItem(in *cart.RemoveItemReq) (*cart.RemoveItemResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"DELETE FROM cart_item WHERE user_id = ? AND product_id = ?",
		in.UserId, in.ProductId,
	)
	if err != nil {
		return nil, err
	}

	return &cart.RemoveItemResp{}, nil
}
