// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeleteReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteReviewLogic {
	return &AdminDeleteReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeleteReviewLogic) AdminDeleteReview(req *types.AdminDeleteReviewReq) (resp *types.OkResp, err error) {
	// todo: add your logic here and delete this line

	return
}
