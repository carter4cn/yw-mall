// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminInjectTrackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminInjectTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminInjectTrackLogic {
	return &AdminInjectTrackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminInjectTrackLogic) AdminInjectTrack(req *types.AdminInjectTrackReq) (resp *types.OkResp, err error) {
	// todo: add your logic here and delete this line

	return
}
