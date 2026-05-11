package logic

import (
	"context"
	"time"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestDeleteReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRequestDeleteReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestDeleteReviewLogic {
	return &RequestDeleteReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RequestDeleteReview lets a merchant submit a takedown request for a review.
func (l *RequestDeleteReviewLogic) RequestDeleteReview(in *review.RequestDeleteReviewReq) (*review.OkResp, error) {
	now := time.Now().Unix()
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO review_delete_request (review_id, shop_id, reason, status, create_time, update_time) VALUES (?, ?, ?, 0, ?, ?)",
		in.ReviewId, in.ShopId, in.Reason, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &review.OkResp{Ok: true}, nil
}
