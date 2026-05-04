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

func (l *SubmitFollowupLogic) SubmitFollowup(req *types.SubmitFollowupReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.ReviewRpc.SubmitFollowup(l.ctx, &reviewpb.SubmitFollowupReq{
		ReviewId: req.ReviewId,
		UserId:   currentUserId(l.ctx),
		Content:  req.Content,
		Media:    reqMediaToProto(req.Media),
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
