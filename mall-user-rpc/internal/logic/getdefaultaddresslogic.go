package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDefaultAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDefaultAddressLogic {
	return &GetDefaultAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDefaultAddressLogic) GetDefaultAddress(in *user.GetDefaultAddressReq) (*user.Address, error) {
	// todo: add your logic here and delete this line

	return &user.Address{}, nil
}
