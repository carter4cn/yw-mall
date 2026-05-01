package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ActivityTemplateModel = (*customActivityTemplateModel)(nil)

type (
	// ActivityTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customActivityTemplateModel.
	ActivityTemplateModel interface {
		activityTemplateModel
	}

	customActivityTemplateModel struct {
		*defaultActivityTemplateModel
	}
)

// NewActivityTemplateModel returns a model for the database table.
func NewActivityTemplateModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ActivityTemplateModel {
	return &customActivityTemplateModel{
		defaultActivityTemplateModel: newActivityTemplateModel(conn, c, opts...),
	}
}
