package logic

import (
	"context"
	"errors"
	"time"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFreightTemplateLogic {
	return &DeleteFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteFreightTemplate soft-deletes by setting status=0 to preserve historical references.
func (l *DeleteFreightTemplateLogic) DeleteFreightTemplate(in *logistics.IdReq) (*logistics.Empty, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE freight_template SET status=0, update_time=? WHERE id=?", now, in.Id); err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
