package logic

import (
	"context"
	"errors"
	"time"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFreightTemplateLogic {
	return &CreateFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateFreightTemplateLogic) CreateFreightTemplate(in *logistics.CreateFreightTemplateReq) (*logistics.CreateFreightTemplateResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	if in.Name == "" {
		return nil, errors.New("name required")
	}
	if in.FirstFee < 0 || in.ExtraFee < 0 {
		return nil, errors.New("fees must be >= 0")
	}
	calc := in.CalcType
	if calc != 1 && calc != 2 {
		calc = 1
	}
	firstVal := in.FirstValue
	if firstVal <= 0 {
		firstVal = 1
	}
	extraVal := in.ExtraValue
	if extraVal <= 0 {
		extraVal = 1
	}
	isDefault := 0
	if in.IsDefault {
		isDefault = 1
	}

	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO freight_template
		 (shop_id, name, calc_type, first_value, first_fee, extra_value, extra_fee, regions, is_default, status, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		in.ShopId, in.Name, calc, firstVal, in.FirstFee, extraVal, in.ExtraFee, in.Regions, isDefault, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &logistics.CreateFreightTemplateResp{Id: id}, nil
}
