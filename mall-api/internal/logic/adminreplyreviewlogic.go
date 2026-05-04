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

type AdminReplyReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminReplyReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminReplyReviewLogic {
	return &AdminReplyReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminReplyReviewLogic) AdminReplyReview(req *types.AdminReplyReviewReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.ReviewRpc.ReplyReview(l.ctx, &reviewpb.ReplyReviewReq{
		ReviewId:       req.Id,
		MerchantUserId: req.MerchantUserId,
		Text:           req.Text,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
