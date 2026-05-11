package logic

import (
	"context"
	"errors"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFreightTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFreightTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFreightTemplatesLogic {
	return &ListFreightTemplatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFreightTemplatesLogic) ListFreightTemplates(in *logistics.ListFreightTemplatesReq) (*logistics.ListFreightTemplatesResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM freight_template WHERE shop_id=? AND status=1", in.ShopId); err != nil {
		return nil, err
	}

	var rows []*freightTemplateRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+freightTemplateCols+" FROM freight_template WHERE shop_id=? AND status=1 ORDER BY is_default DESC, id DESC LIMIT ? OFFSET ?",
		in.ShopId, size, offset); err != nil {
		return nil, err
	}

	out := make([]*logistics.FreightTemplate, 0, len(rows))
	for _, r := range rows {
		out = append(out, toFreightTemplateProto(r))
	}
	return &logistics.ListFreightTemplatesResp{Templates: out, Total: total}, nil
}
