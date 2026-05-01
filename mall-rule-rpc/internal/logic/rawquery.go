package logic

import (
	"context"

	"mall-rule-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// rawQuery and rawCount run direct SQL through a fresh sqlx connection.
// They are used by paged list endpoints since the cached model layer doesn't
// expose generic query methods. Keeping these helpers tiny and local avoids
// surface-area churn in the generated model.

func rawQuery(ctx context.Context, sc *svc.ServiceContext, q string, dst any) error {
	conn := sqlx.NewMysql(sc.Config.DataSource)
	return conn.QueryRowsCtx(ctx, dst, q)
}

func rawCount(ctx context.Context, sc *svc.ServiceContext, q string, dst *int64) error {
	conn := sqlx.NewMysql(sc.Config.DataSource)
	return conn.QueryRowCtx(ctx, dst, q)
}
