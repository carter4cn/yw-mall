// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserReviewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserReviewsLogic {
	return &ListUserReviewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserReviewsLogic) ListUserReviews(req *types.ListUserReviewsReq) (resp *types.ListProductReviewsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
