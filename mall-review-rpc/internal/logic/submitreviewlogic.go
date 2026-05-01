package logic

import (
	"context"
	"strings"

	"mall-common/errorx"
	orderpb "mall-order-rpc/order"
	riskpb "mall-risk-rpc/risk"
	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	orderStatusCompleted    = 3
	rateLimitMaxPerHour     = 10
	rateLimitWindowSeconds  = 3600
	rateLimitSubjectType    = "review_action"
	rateLimitSubjectPrefix  = "submit:"
	blacklistSubjectTypeKey = "user"
)

type SubmitReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReviewLogic {
	return &SubmitReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitReviewLogic) SubmitReview(in *review.SubmitReviewReq) (*review.SubmitReviewResp, error) {
	if !validateScore(in.ScoreMatch) || !validateScore(in.ScoreLogistics) || !validateScore(in.ScoreService) {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	contentLen := len(strings.TrimSpace(in.Content))
	if contentLen < l.svcCtx.Config.Review.ContentMin || contentLen > l.svcCtx.Config.Review.ContentMax {
		return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
	}
	imgs, vids := 0, 0
	for _, m := range in.Media {
		switch m.Type {
		case 1:
			imgs++
		case 2:
			vids++
		default:
			return nil, errorx.NewCodeError(errorx.ReviewMediaInvalid)
		}
		if !validateMediaURL(l.svcCtx.Config.MediaUrlPrefix, m.Url) {
			return nil, errorx.NewCodeError(errorx.ReviewMediaInvalid)
		}
	}
	if imgs > 9 || vids > 1 {
		return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
	}

	item, err := l.svcCtx.OrderRpc.GetOrderItem(l.ctx, &orderpb.GetOrderItemReq{OrderItemId: in.OrderItemId})
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ReviewOrderNotFound)
	}
	if item.UserId != in.UserId {
		return nil, errorx.NewCodeError(errorx.ReviewOrderNotFound)
	}
	if item.OrderStatus != orderStatusCompleted {
		return nil, errorx.NewCodeError(errorx.ReviewOrderNotCompleted)
	}

	if bl, _ := l.svcCtx.RiskRpc.CheckBlacklist(l.ctx, &riskpb.CheckBlacklistReq{
		SubjectType: blacklistSubjectTypeKey, SubjectValue: itoa(in.UserId),
	}); bl != nil && bl.Blacklisted {
		return nil, errorx.NewCodeError(errorx.ReviewRiskBlocked)
	}
	if rl, _ := l.svcCtx.RiskRpc.RateLimit(l.ctx, &riskpb.RateLimitReq{
		SubjectType:   rateLimitSubjectType,
		SubjectValue:  rateLimitSubjectPrefix + itoa(in.UserId),
		MaxCount:      rateLimitMaxPerHour,
		WindowSeconds: rateLimitWindowSeconds,
	}); rl != nil && !rl.Allowed {
		return nil, errorx.NewCodeError(errorx.ReviewRiskBlocked)
	}

	overall := computeOverallScore(in.ScoreMatch, in.ScoreLogistics, in.ScoreService)
	hasMedia := int64(0)
	if len(in.Media) > 0 {
		hasMedia = 1
	}

	var newId int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		ret, err := session.ExecCtx(ctx,
			"INSERT INTO `review`(order_item_id, user_id, product_id, score_overall, score_match, score_logistics, score_service, content, has_media) VALUES (?,?,?,?,?,?,?,?,?)",
			in.OrderItemId, in.UserId, item.ProductId, overall, in.ScoreMatch, in.ScoreLogistics, in.ScoreService, in.Content, hasMedia,
		)
		if err != nil {
			if isDuplicateKey(err) {
				return errorx.NewCodeError(errorx.ReviewAlreadyExists)
			}
			return err
		}
		newId, _ = ret.LastInsertId()
		for i, m := range in.Media {
			sortVal := int32(i)
			if m.Sort != 0 {
				sortVal = m.Sort
			}
			if _, err := session.ExecCtx(ctx,
				"INSERT INTO `review_media`(review_id, media_type, media_url, sort, is_followup) VALUES (?,?,?,?,0)",
				newId, m.Type, m.Url, sortVal,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_, _ = l.svcCtx.Redis.Del(cacheKeyProductSummary(item.ProductId))
	return &review.SubmitReviewResp{ReviewId: newId}, nil
}
