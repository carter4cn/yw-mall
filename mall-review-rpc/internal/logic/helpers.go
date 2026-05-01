package logic

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mall-review-rpc/internal/model"
	"mall-review-rpc/review"

	gosqldriver "github.com/go-sql-driver/mysql"
)

func cacheKeyProductSummary(productId int64) string {
	return fmt.Sprintf("mall:review:summary:%d", productId)
}

func computeOverallScore(scoreMatch, scoreLogistics, scoreService int32) int32 {
	avg := float64(scoreMatch+scoreLogistics+scoreService) / 3.0
	r := int32(avg + 0.5)
	switch {
	case r < 1:
		return 1
	case r > 5:
		return 5
	default:
		return r
	}
}

func validateScore(s int32) bool { return s >= 1 && s <= 5 }

func itoa(i int64) string { return strconv.FormatInt(i, 10) }

func isDuplicateKey(err error) bool {
	var me *gosqldriver.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return true
	}
	return false
}

func validateMediaURL(prefix, url string) bool {
	if prefix == "" {
		return true
	}
	return strings.HasPrefix(url, prefix)
}

func toReviewProto(r *model.Review, media []*model.ReviewMedia) *review.ReviewItem {
	out := &review.ReviewItem{
		Id:             r.Id,
		OrderItemId:    r.OrderItemId,
		UserId:         r.UserId,
		ProductId:      r.ProductId,
		ScoreOverall:   int32(r.ScoreOverall),
		ScoreMatch:     int32(r.ScoreMatch),
		ScoreLogistics: int32(r.ScoreLogistics),
		ScoreService:   int32(r.ScoreService),
		Content:        r.Content,
		Status:         int32(r.Status),
		CreateTime:     r.CreateTime.Unix(),
	}
	if r.FollowupContent.Valid {
		out.FollowupContent = r.FollowupContent.String
	}
	if r.FollowupTime.Valid {
		out.FollowupTime = r.FollowupTime.Time.Unix()
	}
	if r.MerchantReplyText.Valid {
		out.MerchantReplyText = r.MerchantReplyText.String
	}
	if r.MerchantReplyTime.Valid {
		out.MerchantReplyTime = r.MerchantReplyTime.Time.Unix()
	}
	if r.MerchantUserId.Valid {
		out.MerchantUserId = r.MerchantUserId.Int64
	}
	for _, m := range media {
		mm := &review.Media{Type: int32(m.MediaType), Url: m.MediaUrl, Sort: int32(m.Sort)}
		if m.IsFollowup == 1 {
			out.FollowupMedia = append(out.FollowupMedia, mm)
		} else {
			out.Media = append(out.Media, mm)
		}
	}
	return out
}

func groupMediaByReview(media []*model.ReviewMedia) map[int64][]*model.ReviewMedia {
	out := make(map[int64][]*model.ReviewMedia, len(media))
	for _, m := range media {
		out[m.ReviewId] = append(out[m.ReviewId], m)
	}
	return out
}

func clampPaging(page, pageSize int32) (int32, int32, int32) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	return page, pageSize, offset
}
