package logic

import (
	"context"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	// S4.6 always encrypt phone before persisting; empty stays empty.
	phoneEnc, err := cryptox.Encrypt(in.Phone)
	if err != nil {
		return nil, err
	}
	u.Phone = phoneEnc
	u.Avatar = in.Avatar
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		return nil, err
	}

	return &user.UpdateUserResp{}, nil
}
