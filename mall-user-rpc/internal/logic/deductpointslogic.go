package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductPointsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductPointsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductPointsLogic {
	return &DeductPointsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeductPointsLogic) DeductPoints(in *user.DeductPointsReq) (*user.DeductPointsResp, error) {
	// todo: add your logic here and delete this line

	return &user.DeductPointsResp{}, nil
}
