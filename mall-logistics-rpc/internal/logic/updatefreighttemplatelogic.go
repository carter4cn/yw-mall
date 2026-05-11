package logic

import (
	"context"
	"errors"
	"time"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFreightTemplateLogic {
	return &UpdateFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateFreightTemplateLogic) UpdateFreightTemplate(in *logistics.UpdateFreightTemplateReq) (*logistics.Empty, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	if in.FirstFee < 0 || in.ExtraFee < 0 {
		return nil, errors.New("fees must be >= 0")
	}
	isDefault := 0
	if in.IsDefault {
		isDefault = 1
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`UPDATE freight_template SET name=?, first_fee=?, extra_fee=?, is_default=?, update_time=?
		 WHERE id=? AND shop_id=?`,
		in.Name, in.FirstFee, in.ExtraFee, isDefault, now, in.Id, in.ShopId); err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
