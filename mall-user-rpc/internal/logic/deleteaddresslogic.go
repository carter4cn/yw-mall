package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAddressLogic) DeleteAddress(in *user.DeleteAddressReq) (*user.OkResp, error) {
	// todo: add your logic here and delete this line

	return &user.OkResp{}, nil
}
