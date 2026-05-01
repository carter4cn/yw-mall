package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ParticipationRecordModel = (*customParticipationRecordModel)(nil)

type (
	// ParticipationRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customParticipationRecordModel.
	ParticipationRecordModel interface {
		participationRecordModel
	}

	customParticipationRecordModel struct {
		*defaultParticipationRecordModel
	}
)

// NewParticipationRecordModel returns a model for the database table.
func NewParticipationRecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ParticipationRecordModel {
	return &customParticipationRecordModel{
		defaultParticipationRecordModel: newParticipationRecordModel(conn, c, opts...),
	}
}
