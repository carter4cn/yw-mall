package logic

import (
	"context"
	"errors"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrySubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySubscribeLogic {
	return &RetrySubscribeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RetrySubscribeLogic) RetrySubscribe(in *logistics.RetrySubscribeReq) (*logistics.Empty, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errors.New("logistics: shipment not found")
	}
	if err := l.svcCtx.Kuaidi100.Subscribe(l.ctx, s.Carrier, s.TrackingNo); err != nil {
		_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
			"UPDATE `shipment` SET subscribe_status=2 WHERE id=?", s.Id)
		return nil, errors.New("logistics: subscribe failed: " + err.Error())
	}
	_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `shipment` SET subscribe_status=1 WHERE id=?", s.Id)
	return &logistics.Empty{}, nil
}
