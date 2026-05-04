package logic

import (
	"context"

	"mall-common/errorx"
	"mall-review-rpc/internal/model"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReviewLogic {
	return &GetReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReviewLogic) GetReview(in *review.GetReviewReq) (*review.ReviewItem, error) {
	rev, err := l.svcCtx.ReviewModel.FindOne(l.ctx, in.ReviewId)
	if err != nil || rev.Status != 0 {
		return nil, errorx.NewCodeError(errorx.ReviewNotFound)
	}
	media, err := fetchMediaByReviewIds(l.ctx, l.svcCtx, []int64{rev.Id})
	if err != nil {
		return nil, err
	}
	return toReviewProto(rev, media), nil
}

func fetchMediaByReviewIds(ctx context.Context, svcCtx *svc.ServiceContext, ids []int64) ([]*model.ReviewMedia, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := ""
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	q := "SELECT id, review_id, media_type, media_url, sort, is_followup, create_time FROM review_media WHERE review_id IN (" + placeholders + ") ORDER BY review_id, is_followup, sort"
	var rows []*model.ReviewMedia
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}
