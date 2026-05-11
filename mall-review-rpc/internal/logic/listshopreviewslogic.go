package logic

import (
	"context"

	"mall-review-rpc/internal/model"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopReviewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopReviewsLogic {
	return &ListShopReviewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListShopReviews returns paginated reviews scoped to a shop. Uses the
// review.shop_id column populated at submit time (P1 admin migration).
func (l *ListShopReviewsLogic) ListShopReviews(in *review.ListShopReviewsReq) (*review.ListProductReviewsResp, error) {
	conds := "WHERE shop_id = ? AND status = 0"
	args := []any{in.ShopId}
	if in.Score >= 1 && in.Score <= 5 {
		conds += " AND score_overall = ?"
		args = append(args, in.Score)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM review "+conds, args...); err != nil {
		return nil, err
	}

	_, pageSize, offset := clampPaging(in.Page, in.PageSize)
	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, pageSize, offset)

	rows := []*model.Review{}
	q := "SELECT " + reviewSelectCols + " FROM review " + conds + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, pagedArgs...); err != nil {
		return nil, err
	}

	return assembleListResp(l.ctx, l.svcCtx, rows, total)
}
