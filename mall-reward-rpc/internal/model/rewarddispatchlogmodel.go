package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RewardDispatchLogModel = (*customRewardDispatchLogModel)(nil)

type (
	// RewardDispatchLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRewardDispatchLogModel.
	RewardDispatchLogModel interface {
		rewardDispatchLogModel
	}

	customRewardDispatchLogModel struct {
		*defaultRewardDispatchLogModel
	}
)

// NewRewardDispatchLogModel returns a model for the database table.
func NewRewardDispatchLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RewardDispatchLogModel {
	return &customRewardDispatchLogModel{
		defaultRewardDispatchLogModel: newRewardDispatchLogModel(conn, c, opts...),
	}
}
