package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/types"
	reviewpb "mall-review-rpc/review"
)

func currentUserId(ctx context.Context) int64 {
	return middleware.UidFromCtx(ctx)
}

func protoMediaList(in []*reviewpb.Media) []types.ReviewMediaItem {
	out := make([]types.ReviewMediaItem, 0, len(in))
	for _, m := range in {
		out = append(out, types.ReviewMediaItem{Type: m.Type, Url: m.Url, Sort: m.Sort})
	}
	return out
}

func protoReviewToType(r *reviewpb.ReviewItem) types.ReviewItem {
	return types.ReviewItem{
		Id:                r.Id,
		OrderItemId:       r.OrderItemId,
		UserId:            r.UserId,
		ProductId:         r.ProductId,
		ScoreOverall:      r.ScoreOverall,
		ScoreMatch:        r.ScoreMatch,
		ScoreLogistics:    r.ScoreLogistics,
		ScoreService:      r.ScoreService,
		Content:           r.Content,
		Media:             protoMediaList(r.Media),
		FollowupContent:   r.FollowupContent,
		FollowupTime:      r.FollowupTime,
		FollowupMedia:     protoMediaList(r.FollowupMedia),
		MerchantReplyText: r.MerchantReplyText,
		MerchantReplyTime: r.MerchantReplyTime,
		MerchantUserId:    r.MerchantUserId,
		CreateTime:        r.CreateTime,
	}
}

func reqMediaToProto(in []types.ReviewMediaItem) []*reviewpb.Media {
	out := make([]*reviewpb.Media, 0, len(in))
	for i, m := range in {
		sort := m.Sort
		if sort == 0 {
			sort = int32(i)
		}
		out = append(out, &reviewpb.Media{Type: m.Type, Url: m.Url, Sort: sort})
	}
	return out
}
