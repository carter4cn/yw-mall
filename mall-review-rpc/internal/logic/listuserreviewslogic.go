package logic

import (
	"context"

	"mall-review-rpc/internal/model"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserReviewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserReviewsLogic {
	return &ListUserReviewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserReviewsLogic) ListUserReviews(in *review.ListUserReviewsReq) (*review.ListProductReviewsResp, error) {
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM review WHERE user_id=? AND status=0", in.UserId); err != nil {
		return nil, err
	}

	_, pageSize, offset := clampPaging(in.Page, in.PageSize)
	rows := []*model.Review{}
	q := "SELECT " + reviewSelectCols + " FROM review WHERE user_id=? AND status=0 ORDER BY create_time DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, in.UserId, pageSize, offset); err != nil {
		return nil, err
	}

	return assembleListResp(l.ctx, l.svcCtx, rows, total)
}
