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

func (l *GetRatingSummaryLogic) GetRatingSummary(req *types.GetRatingSummaryReq) (*types.GetRatingSummaryResp, error) {
	s, err := l.svcCtx.ReviewRpc.GetProductRatingSummary(l.ctx, &reviewpb.GetProductRatingSummaryReq{ProductId: req.ProductId})
	if err != nil {
		return nil, err
	}
	return &types.GetRatingSummaryResp{
		Avg:            s.Avg,
		Count:          s.Count,
		Distribution:   s.Distribution,
		WithMediaCount: s.WithMediaCount,
	}, nil
}
