package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RuleEvaluationLogModel = (*customRuleEvaluationLogModel)(nil)

type (
	// RuleEvaluationLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRuleEvaluationLogModel.
	RuleEvaluationLogModel interface {
		ruleEvaluationLogModel
	}

	customRuleEvaluationLogModel struct {
		*defaultRuleEvaluationLogModel
	}
)

// NewRuleEvaluationLogModel returns a model for the database table.
func NewRuleEvaluationLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RuleEvaluationLogModel {
	return &customRuleEvaluationLogModel{
		defaultRuleEvaluationLogModel: newRuleEvaluationLogModel(conn, c, opts...),
	}
}
