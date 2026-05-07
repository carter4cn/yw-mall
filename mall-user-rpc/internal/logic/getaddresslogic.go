package logic

import (
	"context"

	"mall-common/errorx"
	"mall-user-rpc/internal/model"
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
	return toAddrProto(addr), nil
}
