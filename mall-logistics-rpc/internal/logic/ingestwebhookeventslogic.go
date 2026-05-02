package logic

import (
	"context"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type IngestWebhookEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIngestWebhookEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IngestWebhookEventsLogic {
	return &IngestWebhookEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IngestWebhookEventsLogic) IngestWebhookEvents(in *logistics.IngestWebhookEventsReq) (*logistics.Empty, error) {
	// todo: add your logic here and delete this line

	return &logistics.Empty{}, nil
}
