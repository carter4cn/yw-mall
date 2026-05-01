package logic

import (
	"context"
	"fmt"

	"mall-workflow-rpc/internal/fsm"
	"mall-workflow-rpc/internal/model"
	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterDefinitionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterDefinitionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterDefinitionLogic {
	return &RegisterDefinitionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterDefinitionLogic) RegisterDefinition(in *workflow.RegisterDefinitionReq) (*workflow.RegisterDefinitionResp, error) {
	if _, err := fsm.Parse(&model.WorkflowDefinition{
		Code:            in.Code,
		Description:     in.Description,
		StatesJson:      in.StatesJson,
		TransitionsJson: in.TransitionsJson,
	}); err != nil {
		return nil, fmt.Errorf("invalid definition: %w", err)
	}

	existing, err := l.svcCtx.WorkflowDefinitionModel.FindOneByCode(l.ctx, in.Code)
	if err == nil && existing != nil {
		existing.Description = in.Description
		existing.StatesJson = in.StatesJson
		existing.TransitionsJson = in.TransitionsJson
		existing.Version++
		if uerr := l.svcCtx.WorkflowDefinitionModel.Update(l.ctx, existing); uerr != nil {
			return nil, uerr
		}
		l.svcCtx.Registry.Bust(existing.Id)
		return &workflow.RegisterDefinitionResp{Id: int64(existing.Id)}, nil
	}

	res, err := l.svcCtx.WorkflowDefinitionModel.Insert(l.ctx, &model.WorkflowDefinition{
		Code:            in.Code,
		Description:     in.Description,
		StatesJson:      in.StatesJson,
		TransitionsJson: in.TransitionsJson,
		Version:         1,
		Status:          "ACTIVE",
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &workflow.RegisterDefinitionResp{Id: id}, nil
}
