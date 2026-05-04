// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	reviewpb "mall-review-rpc/review"

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

func (l *AdminDeleteReviewLogic) AdminDeleteReview(req *types.AdminDeleteReviewReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.ReviewRpc.SoftDeleteReview(l.ctx, &reviewpb.SoftDeleteReviewReq{ReviewId: req.Id}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
