package logic

import (
	"context"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type InjectTrackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInjectTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InjectTrackLogic {
	return &InjectTrackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InjectTrackLogic) InjectTrack(in *logistics.InjectTrackReq) (*logistics.Empty, error) {
	// todo: add your logic here and delete this line

	return &logistics.Empty{}, nil
}
