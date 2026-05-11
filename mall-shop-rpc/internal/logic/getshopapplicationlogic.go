package logic

import (
	"context"
	"errors"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetShopApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShopApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShopApplicationLogic {
	return &GetShopApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetShopApplicationLogic) GetShopApplication(in *shop.GetShopApplicationReq) (*shop.ShopApplication, error) {
	var r applicationRow
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &r,
		"SELECT "+applicationCols+" FROM shop_application WHERE id=? LIMIT 1", in.Id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("application not found")
		}
		return nil, err
	}
	return toApplicationProto(&r), nil
}
