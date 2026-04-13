package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearCartLogic) ClearCart(in *cart.ClearCartReq) (*cart.ClearCartResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"DELETE FROM cart_item WHERE user_id = ?",
		in.UserId,
	)
	if err != nil {
		return nil, err
	}

	return &cart.ClearCartResp{}, nil
}
