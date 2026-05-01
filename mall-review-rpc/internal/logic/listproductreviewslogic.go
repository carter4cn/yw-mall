package logic

import (
	"context"
	"strings"

	"mall-review-rpc/internal/model"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

const reviewSelectCols = "id, order_item_id, user_id, product_id, score_overall, score_match, score_logistics, score_service, content, has_media, followup_content, followup_time, merchant_reply_text, merchant_reply_time, merchant_user_id, status, create_time, update_time"

type ListProductReviewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductReviewsLogic {
	return &ListProductReviewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductReviewsLogic) ListProductReviews(in *review.ListProductReviewsReq) (*review.ListProductReviewsResp, error) {
	conds := []string{"product_id = ?", "status = 0"}
	args := []any{in.ProductId}
	if in.Score >= 1 && in.Score <= 5 {
		conds = append(conds, "score_overall = ?")
		args = append(args, in.Score)
	}
	if in.WithMedia {
		conds = append(conds, "has_media = 1")
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	orderBy := "ORDER BY create_time DESC"
	switch in.Sort {
	case "score":
		orderBy = "ORDER BY score_overall DESC, create_time DESC"
	case "hasMedia":
		orderBy = "ORDER BY has_media DESC, create_time DESC"
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM review "+where, args...); err != nil {
		return nil, err
	}

	_, pageSize, offset := clampPaging(in.Page, in.PageSize)
	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, pageSize, offset)

	rows := []*model.Review{}
	q := "SELECT " + reviewSelectCols + " FROM review " + where + " " + orderBy + " LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, pagedArgs...); err != nil {
		return nil, err
	}

	return assembleListResp(l.ctx, l.svcCtx, rows, total)
}

func assembleListResp(ctx context.Context, svcCtx *svc.ServiceContext, rows []*model.Review, total int64) (*review.ListProductReviewsResp, error) {
	if len(rows) == 0 {
		return &review.ListProductReviewsResp{Reviews: nil, Total: total}, nil
	}
	ids := make([]int64, len(rows))
	for i, r := range rows {
		ids[i] = r.Id
	}
	media, err := fetchMediaByReviewIds(ctx, svcCtx, ids)
	if err != nil {
		return nil, err
	}
	mediaByReview := groupMediaByReview(media)
	out := make([]*review.ReviewItem, len(rows))
	for i, r := range rows {
		out[i] = toReviewProto(r, mediaByReview[r.Id])
	}
	return &review.ListProductReviewsResp{Reviews: out, Total: total}, nil
}
