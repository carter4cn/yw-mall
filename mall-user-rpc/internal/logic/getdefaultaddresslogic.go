package logic

import (
	"context"

	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	var addr model.UserAddress
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &addr,
		"SELECT * FROM user_address WHERE user_id=? AND is_default=1 LIMIT 1", in.UserId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &user.Address{Id: 0}, nil
		}
		return nil, err
	}
	return toAddrProto(&addr), nil
}
