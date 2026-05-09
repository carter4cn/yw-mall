package logic

import (
	"context"

	"mall-user-rpc/internal/model"
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
	var rows []*model.UserAddress
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT * FROM user_address WHERE user_id=? ORDER BY is_default DESC, update_time DESC", in.UserId)
	if err != nil {
		return nil, err
	}
	out := make([]*user.Address, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAddrProto(a))
	}
	return &user.ListAddressesResp{Addresses: out}, nil
}
