package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ReviewModel = (*customReviewModel)(nil)

type (
	// ReviewModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReviewModel.
	ReviewModel interface {
		reviewModel
	}

	customReviewModel struct {
		*defaultReviewModel
	}
)

// NewReviewModel returns a model for the database table.
func NewReviewModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ReviewModel {
	return &customReviewModel{
		defaultReviewModel: newReviewModel(conn, c, opts...),
	}
}
