// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitFollowupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitFollowupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitFollowupLogic {
	return &SubmitFollowupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitFollowupLogic) SubmitFollowup(req *types.SubmitFollowupReq) (resp *types.OkResp, err error) {
	// todo: add your logic here and delete this line

	return
}
