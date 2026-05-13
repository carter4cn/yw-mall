package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultAddressLogic) SetDefaultAddress(req *types.SetDefaultAddressReq) (*types.FollowShopResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	_, err := l.svcCtx.UserRpc.SetDefaultAddress(l.ctx, &userclient.SetDefaultAddressReq{
		UserId: userId,
		Id:     req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.FollowShopResp{Ok: true}, nil
}
