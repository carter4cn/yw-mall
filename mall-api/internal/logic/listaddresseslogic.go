package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAddressesLogic) ListAddresses() (*types.ListAddressesResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.UserRpc.ListAddresses(l.ctx, &userclient.ListAddressesReq{UserId: userId})
	if err != nil {
		return nil, err
	}

	addrs := make([]types.AddressItem, 0, len(res.Addresses))
	for _, a := range res.Addresses {
		addrs = append(addrs, types.AddressItem{
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
		})
	}
	return &types.ListAddressesResp{Addresses: addrs}, nil
}
