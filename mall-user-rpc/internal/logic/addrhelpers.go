package logic

import (
	"mall-user-rpc/internal/model"
	"mall-user-rpc/user"
)

func toAddrProto(a *model.UserAddress) *user.Address {
	return &user.Address{
		Id:           int64(a.Id),
		UserId:       int64(a.UserId),
		ReceiverName: a.ReceiverName,
		Phone:        a.Phone,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		Detail:       a.Detail,
		IsDefault:    a.IsDefault == 1,
		CreateTime:   a.CreateTime,
	}
}
