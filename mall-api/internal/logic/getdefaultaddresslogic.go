package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDefaultAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDefaultAddressLogic {
	return &GetDefaultAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDefaultAddressLogic) GetDefaultAddress() (*types.AddressItem, error) {
	userId := middleware.UidFromCtx(l.ctx)

	a, err := l.svcCtx.UserRpc.GetDefaultAddress(l.ctx, &userclient.GetDefaultAddressReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &types.AddressItem{
		Id:           a.Id,
		UserId:       a.UserId,
		ReceiverName: a.ReceiverName,
		Phone:        a.Phone,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		Detail:       a.Detail,
		IsDefault:    a.IsDefault,
		CreateTime:   a.CreateTime,
	}, nil
}
