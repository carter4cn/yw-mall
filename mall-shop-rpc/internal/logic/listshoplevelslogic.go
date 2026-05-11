package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShopLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShopLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShopLevelsLogic {
	return &ListShopLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListShopLevelsLogic) ListShopLevels(_ *shop.Empty) (*shop.ListShopLevelsResp, error) {
	var rows []*levelTemplateRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+levelTemplateCols+" FROM shop_level_template ORDER BY level ASC"); err != nil {
		return nil, err
	}
	out := make([]*shop.ShopLevelTemplate, 0, len(rows))
	for _, r := range rows {
		out = append(out, toLevelTemplateProto(r))
	}
	return &shop.ListShopLevelsResp{Levels: out}, nil
}
