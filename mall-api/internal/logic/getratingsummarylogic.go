// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRatingSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRatingSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRatingSummaryLogic {
	return &GetRatingSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRatingSummaryLogic) GetRatingSummary(req *types.GetRatingSummaryReq) (resp *types.GetRatingSummaryResp, err error) {
	// todo: add your logic here and delete this line

	return
}
