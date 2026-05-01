package logic

import (
	"context"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDefinitionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDefinitionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDefinitionLogic {
	return &GetDefinitionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDefinitionLogic) GetDefinition(in *workflow.IdReq) (*workflow.WorkflowDefinition, error) {
	d, err := l.svcCtx.WorkflowDefinitionModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &workflow.WorkflowDefinition{
		Id:              int64(d.Id),
		Code:            d.Code,
		Description:     d.Description,
		StatesJson:      d.StatesJson,
		TransitionsJson: d.TransitionsJson,
		Version:         int32(d.Version),
		Status:          d.Status,
	}, nil
}
