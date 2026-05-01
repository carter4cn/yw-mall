package logic

import (
	"context"
	"database/sql"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/model"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateActivityLogic) CreateActivity(in *activity.CreateActivityReq) (*activity.CreateActivityResp, error) {
	res, err := l.svcCtx.ActivityModel.Insert(l.ctx, &model.Activity{
		Code:                 in.Code,
		Title:                in.Title,
		Description:          sql.NullString{String: in.Description, Valid: in.Description != ""},
		Type:                 in.Type,
		Status:               "DRAFT",
		StartTime:            time.Unix(in.StartTime, 0),
		EndTime:              time.Unix(in.EndTime, 0),
		TemplateId:           in.TemplateId,
		RuleSetId:            in.RuleSetId,
		WorkflowDefinitionId: in.WorkflowDefinitionId,
		ConfigJson:           sql.NullString{String: in.ConfigJson, Valid: in.ConfigJson != ""},
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &activity.CreateActivityResp{Id: id}, nil
}
