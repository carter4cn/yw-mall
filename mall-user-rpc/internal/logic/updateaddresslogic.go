package logic

import (
	"context"
	"time"

	"mall-common/errorx"
	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAddressLogic) UpdateAddress(in *user.UpdateAddressReq) (*user.OkResp, error) {
	addr, err := l.svcCtx.UserAddressModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.UserAddressNotFound)
		}
		return nil, err
	}
	if int64(addr.UserId) != in.UserId {
		return nil, errorx.NewCodeError(errorx.UserAddressForbidden)
	}
	if in.ReceiverName != "" {
		addr.ReceiverName = in.ReceiverName
	}
	if in.Phone != "" {
		addr.Phone = in.Phone
	}
	if in.Province != "" {
		addr.Province = in.Province
	}
	if in.City != "" {
		addr.City = in.City
	}
	if in.District != "" {
		addr.District = in.District
	}
	if in.Detail != "" {
		addr.Detail = in.Detail
	}
	addr.UpdateTime = time.Now().Unix()
	if err := l.svcCtx.UserAddressModel.Update(l.ctx, addr); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}
