package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterTemplateLogic {
	return &RegisterTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterTemplateLogic) RegisterTemplate(in *activity.RegisterTemplateReq) (*activity.Empty, error) {
	// todo: add your logic here and delete this line

	return &activity.Empty{}, nil
}
