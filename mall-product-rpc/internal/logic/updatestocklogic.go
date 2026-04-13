package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStockLogic {
	return &UpdateStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateStockLogic) UpdateStock(in *product.UpdateStockReq) (*product.UpdateStockResp, error) {
	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)

	query := "UPDATE product SET stock = stock + ? WHERE id = ? AND stock + ? >= 0"
	result, err := conn.ExecCtx(l.ctx, query, in.Delta, in.Id, in.Delta)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, errors.New("stock not enough")
	}

	return &product.UpdateStockResp{}, nil
}
