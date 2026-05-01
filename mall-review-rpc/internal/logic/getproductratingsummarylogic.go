package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductRatingSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductRatingSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductRatingSummaryLogic {
	return &GetProductRatingSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductRatingSummaryLogic) GetProductRatingSummary(in *review.GetProductRatingSummaryReq) (*review.RatingSummary, error) {
	// todo: add your logic here and delete this line

	return &review.RatingSummary{}, nil
}
