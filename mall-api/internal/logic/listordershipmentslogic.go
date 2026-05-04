// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrderShipmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrderShipmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderShipmentsLogic {
	return &ListOrderShipmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrderShipmentsLogic) ListOrderShipments(req *types.ListOrderShipmentsReq) (resp *types.ListOrderShipmentsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
