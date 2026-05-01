package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ParticipationTokenModel = (*customParticipationTokenModel)(nil)

type (
	// ParticipationTokenModel is an interface to be customized, add more methods here,
	// and implement the added methods in customParticipationTokenModel.
	ParticipationTokenModel interface {
		participationTokenModel
	}

	customParticipationTokenModel struct {
		*defaultParticipationTokenModel
	}
)

// NewParticipationTokenModel returns a model for the database table.
func NewParticipationTokenModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ParticipationTokenModel {
	return &customParticipationTokenModel{
		defaultParticipationTokenModel: newParticipationTokenModel(conn, c, opts...),
	}
}
