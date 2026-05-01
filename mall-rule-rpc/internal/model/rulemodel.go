package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RuleModel = (*customRuleModel)(nil)

type (
	// RuleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRuleModel.
	RuleModel interface {
		ruleModel
	}

	customRuleModel struct {
		*defaultRuleModel
	}
)

// NewRuleModel returns a model for the database table.
func NewRuleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RuleModel {
	return &customRuleModel{
		defaultRuleModel: newRuleModel(conn, c, opts...),
	}
}
