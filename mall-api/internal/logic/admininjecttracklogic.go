// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"

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

func (l *AdminInjectTrackLogic) AdminInjectTrack(req *types.AdminInjectTrackReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.LogisticsRpc.InjectTrack(l.ctx, &logisticspb.InjectTrackReq{
		ShipmentId:    req.Id,
		StateInternal: req.StateInternal,
		Location:      req.Location,
		Description:   req.Description,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
