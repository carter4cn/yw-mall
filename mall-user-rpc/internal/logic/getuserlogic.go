package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *user.GetUserReq) (*user.GetUserResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	return &user.GetUserResp{
		Id:         int64(u.Id),
		Username:   u.Username,
		Phone:      u.Phone,
		Avatar:     u.Avatar,
		CreateTime: u.CreateTime.Unix(),
	}, nil
}
