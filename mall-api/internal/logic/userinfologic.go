// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResp, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.UserRpc.GetUser(l.ctx, &user.GetUserReq{
		Id: userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.UserInfoResp{
		Id:         res.Id,
		Username:   res.Username,
		Phone:      res.Phone,
		Avatar:     res.Avatar,
		CreateTime: res.CreateTime,
	}, nil
}
