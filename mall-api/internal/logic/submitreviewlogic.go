// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReviewLogic {
	return &SubmitReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReviewLogic) SubmitReview(req *types.SubmitReviewReq) (resp *types.SubmitReviewResp, err error) {
	// todo: add your logic here and delete this line

	return
}
