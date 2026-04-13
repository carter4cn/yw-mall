package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PaymentModel = (*customPaymentModel)(nil)

type (
	// PaymentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPaymentModel.
	PaymentModel interface {
		paymentModel
	}

	customPaymentModel struct {
		*defaultPaymentModel
	}
)

// NewPaymentModel returns a model for the database table.
func NewPaymentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PaymentModel {
	return &customPaymentModel{
		defaultPaymentModel: newPaymentModel(conn, c, opts...),
	}
}
