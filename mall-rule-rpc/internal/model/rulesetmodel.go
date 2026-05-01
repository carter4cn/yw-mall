package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RuleSetModel = (*customRuleSetModel)(nil)

type (
	// RuleSetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRuleSetModel.
	RuleSetModel interface {
		ruleSetModel
	}

	customRuleSetModel struct {
		*defaultRuleSetModel
	}
)

// NewRuleSetModel returns a model for the database table.
func NewRuleSetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RuleSetModel {
	return &customRuleSetModel{
		defaultRuleSetModel: newRuleSetModel(conn, c, opts...),
	}
}
