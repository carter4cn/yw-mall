package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAddressLogic) GetAddress(in *user.GetAddressReq) (*user.Address, error) {
	// todo: add your logic here and delete this line

	return &user.Address{}, nil
}
