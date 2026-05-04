package logic

import (
	"context"
	"strings"
	"time"

	"mall-common/errorx"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SubmitFollowupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitFollowupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitFollowupLogic {
	return &SubmitFollowupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitFollowupLogic) SubmitFollowup(in *review.SubmitFollowupReq) (*review.Empty, error) {
	contentLen := len(strings.TrimSpace(in.Content))
	if contentLen < l.svcCtx.Config.Review.ContentMin || contentLen > l.svcCtx.Config.Followup.MaxLength {
		return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
	}
	for _, m := range in.Media {
		if m.Type != 1 && m.Type != 2 {
			return nil, errorx.NewCodeError(errorx.ReviewMediaInvalid)
		}
		if !validateMediaURL(l.svcCtx.Config.MediaUrlPrefix, m.Url) {
			return nil, errorx.NewCodeError(errorx.ReviewMediaInvalid)
		}
	}

	rev, err := l.svcCtx.ReviewModel.FindOne(l.ctx, in.ReviewId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ReviewNotFound)
	}
	if rev.UserId != in.UserId || rev.Status != 0 {
		return nil, errorx.NewCodeError(errorx.ReviewNotFound)
	}
	if rev.FollowupContent.Valid {
		return nil, errorx.NewCodeError(errorx.ReviewFollowupNotAllowed)
	}
	minDelay := time.Duration(l.svcCtx.Config.Followup.MinDelayDays) * 24 * time.Hour
	if time.Since(rev.CreateTime) < minDelay {
		return nil, errorx.NewCodeError(errorx.ReviewFollowupNotAllowed)
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		res, err := session.ExecCtx(ctx,
			"UPDATE `review` SET followup_content=?, followup_time=NOW() WHERE id=? AND followup_content IS NULL",
			in.Content, in.ReviewId,
		)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return errorx.NewCodeError(errorx.ReviewFollowupNotAllowed)
		}
		for i, m := range in.Media {
			sortVal := int32(i)
			if m.Sort != 0 {
				sortVal = m.Sort
			}
			if _, err := session.ExecCtx(ctx,
				"INSERT INTO `review_media`(review_id, media_type, media_url, sort, is_followup) VALUES (?,?,?,?,1)",
				in.ReviewId, m.Type, m.Url, sortVal,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_, _ = l.svcCtx.Redis.Del(cacheKeyProductSummary(rev.ProductId))
	return &review.Empty{}, nil
}
