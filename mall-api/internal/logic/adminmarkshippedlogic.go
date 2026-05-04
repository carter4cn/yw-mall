// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminMarkShippedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminMarkShippedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminMarkShippedLogic {
	return &AdminMarkShippedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminMarkShippedLogic) AdminMarkShipped(req *types.AdminMarkShippedReq) (resp *types.OkResp, err error) {
	// todo: add your logic here and delete this line

	return
}
