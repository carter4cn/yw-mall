package logic

import (
	"context"
	"encoding/json"

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

type cachedSummary struct {
	Avg            float64         `json:"avg"`
	Count          int64           `json:"count"`
	Distribution   map[int32]int64 `json:"distribution"`
	WithMediaCount int64           `json:"withMediaCount"`
}

func (l *GetProductRatingSummaryLogic) GetProductRatingSummary(in *review.GetProductRatingSummaryReq) (*review.RatingSummary, error) {
	key := cacheKeyProductSummary(in.ProductId)
	if v, err := l.svcCtx.Redis.GetCtx(l.ctx, key); err == nil && v != "" {
		var c cachedSummary
		if err := json.Unmarshal([]byte(v), &c); err == nil {
			return &review.RatingSummary{
				Avg: c.Avg, Count: c.Count, Distribution: c.Distribution, WithMediaCount: c.WithMediaCount,
			}, nil
		}
	}

	type aggRow struct {
		ScoreOverall int64 `db:"score_overall"`
		Count        int64 `db:"count"`
		WithMedia    int64 `db:"with_media"`
	}
	rows := []*aggRow{}
	q := `SELECT score_overall, COUNT(*) AS count, SUM(has_media) AS with_media
	      FROM review WHERE product_id=? AND status=0 GROUP BY score_overall`
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, in.ProductId); err != nil {
		return nil, err
	}

	dist := map[int32]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	var sum, count, withMedia int64
	for _, r := range rows {
		score := int32(r.ScoreOverall)
		if score >= 1 && score <= 5 {
			dist[score] = r.Count
		}
		count += r.Count
		withMedia += r.WithMedia
		sum += r.ScoreOverall * r.Count
	}
	var avg float64
	if count > 0 {
		avg = float64(sum) / float64(count)
	}

	resp := &review.RatingSummary{
		Avg: avg, Count: count, Distribution: dist, WithMediaCount: withMedia,
	}

	ttl := l.svcCtx.Config.CacheTTLSeconds
	if ttl < 60 {
		ttl = 300
	}
	if b, err := json.Marshal(cachedSummary{Avg: avg, Count: count, Distribution: dist, WithMediaCount: withMedia}); err == nil {
		_ = l.svcCtx.Redis.SetexCtx(l.ctx, key, string(b), ttl)
	}
	return resp, nil
}
