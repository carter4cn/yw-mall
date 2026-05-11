package logic

import (
	"context"
	"errors"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFreightTemplateLogic {
	return &GetFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFreightTemplateLogic) GetFreightTemplate(in *logistics.IdReq) (*logistics.FreightTemplate, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}
	var row freightTemplateRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT "+freightTemplateCols+" FROM freight_template WHERE id=? LIMIT 1", in.Id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return toFreightTemplateProto(&row), nil
}
