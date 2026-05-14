package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListItemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListItemsLogic {
	return &ListItemsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type cartItemRow struct {
	ProductId int64
	Quantity  int32
	Selected  int8
}

func (l *ListItemsLogic) ListItems(in *cart.ListItemsReq) (*cart.ListItemsResp, error) {
	// ProxySQL routes plain SELECTs to read replicas (hostgroup 20) which lag
	// behind master by 1-2 s. Wrapping the read in a transaction pins it to
	// the default hostgroup (10 = master), giving us read-after-write freshness
	// — critical for cart UX where the user expects an item to appear in the
	// list immediately after `Add`.
	var rows []cartItemRow
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		return sess.QueryRowsCtx(ctx, &rows,
			"SELECT product_id, quantity, selected FROM cart_item WHERE user_id = ?",
			in.UserId,
		)
	})
	if err != nil {
		return nil, err
	}

	items := make([]*cart.CartItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &cart.CartItem{
			ProductId: row.ProductId,
			Quantity:  row.Quantity,
			Selected:  row.Selected == 1,
		})
	}

	return &cart.ListItemsResp{Items: items}, nil
}
