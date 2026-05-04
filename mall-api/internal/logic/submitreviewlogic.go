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

func (l *SubmitReviewLogic) SubmitReview(req *types.SubmitReviewReq) (*types.SubmitReviewResp, error) {
	resp, err := l.svcCtx.ReviewRpc.SubmitReview(l.ctx, &reviewpb.SubmitReviewReq{
		OrderItemId:    req.OrderItemId,
		UserId:         currentUserId(l.ctx),
		ScoreMatch:     req.ScoreMatch,
		ScoreLogistics: req.ScoreLogistics,
		ScoreService:   req.ScoreService,
		Content:        req.Content,
		Media:          reqMediaToProto(req.Media),
	})
	if err != nil {
		return nil, err
	}
	return &types.SubmitReviewResp{ReviewId: resp.ReviewId}, nil
}
