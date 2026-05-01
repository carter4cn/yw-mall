package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RewardRecordModel = (*customRewardRecordModel)(nil)

type (
	// RewardRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRewardRecordModel.
	RewardRecordModel interface {
		rewardRecordModel
	}

	customRewardRecordModel struct {
		*defaultRewardRecordModel
	}
)

// NewRewardRecordModel returns a model for the database table.
func NewRewardRecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RewardRecordModel {
	return &customRewardRecordModel{
		defaultRewardRecordModel: newRewardRecordModel(conn, c, opts...),
	}
}
