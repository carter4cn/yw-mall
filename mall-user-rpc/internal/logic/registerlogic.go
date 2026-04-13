package logic

import (
	"context"

	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username: in.Username,
		Password: string(hash),
		Phone:    in.Phone,
	})
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &user.RegisterResp{Id: id}, nil
}
