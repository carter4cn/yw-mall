package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RiskScoreCacheModel = (*customRiskScoreCacheModel)(nil)

type (
	// RiskScoreCacheModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRiskScoreCacheModel.
	RiskScoreCacheModel interface {
		riskScoreCacheModel
	}

	customRiskScoreCacheModel struct {
		*defaultRiskScoreCacheModel
	}
)

// NewRiskScoreCacheModel returns a model for the database table.
func NewRiskScoreCacheModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RiskScoreCacheModel {
	return &customRiskScoreCacheModel{
		defaultRiskScoreCacheModel: newRiskScoreCacheModel(conn, c, opts...),
	}
}
