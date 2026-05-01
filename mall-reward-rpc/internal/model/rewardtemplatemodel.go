package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RewardTemplateModel = (*customRewardTemplateModel)(nil)

type (
	// RewardTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRewardTemplateModel.
	RewardTemplateModel interface {
		rewardTemplateModel
	}

	customRewardTemplateModel struct {
		*defaultRewardTemplateModel
	}
)

// NewRewardTemplateModel returns a model for the database table.
func NewRewardTemplateModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RewardTemplateModel {
	return &customRewardTemplateModel{
		defaultRewardTemplateModel: newRewardTemplateModel(conn, c, opts...),
	}
}
