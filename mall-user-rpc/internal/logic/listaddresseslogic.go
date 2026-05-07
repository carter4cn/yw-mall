package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAddressesLogic) ListAddresses(in *user.ListAddressesReq) (*user.ListAddressesResp, error) {
	// todo: add your logic here and delete this line

	return &user.ListAddressesResp{}, nil
}
