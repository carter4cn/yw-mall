package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ReviewMediaModel = (*customReviewMediaModel)(nil)

type (
	// ReviewMediaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReviewMediaModel.
	ReviewMediaModel interface {
		reviewMediaModel
	}

	customReviewMediaModel struct {
		*defaultReviewMediaModel
	}
)

// NewReviewMediaModel returns a model for the database table.
func NewReviewMediaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ReviewMediaModel {
	return &customReviewMediaModel{
		defaultReviewMediaModel: newReviewMediaModel(conn, c, opts...),
	}
}
