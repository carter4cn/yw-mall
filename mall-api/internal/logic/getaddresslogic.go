package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAddressLogic) GetAddress(req *types.GetAddressReq) (*types.AddressItem, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	a, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &userclient.GetAddressReq{
		UserId: userId,
		Id:     req.Id,
	})
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
