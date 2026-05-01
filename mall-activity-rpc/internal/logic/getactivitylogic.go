package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityLogic {
	return &GetActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityLogic) GetActivity(in *activity.IdReq) (*activity.ActivityDetail, error) {
	a, err := l.svcCtx.ActivityModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	info := &activity.ActivityInfo{
		Id:                   int64(a.Id),
		Code:                 a.Code,
		Title:                a.Title,
		Description:          a.Description.String,
		Type:                 a.Type,
		Status:               a.Status,
		StartTime:            a.StartTime.Unix(),
		EndTime:              a.EndTime.Unix(),
		RuleSetId:            a.RuleSetId,
		WorkflowDefinitionId: a.WorkflowDefinitionId,
		TemplateId:           a.TemplateId,
		ConfigJson:           a.ConfigJson.String,
		CreateTime:           a.CreateTime.Unix(),
		UpdateTime:           a.UpdateTime.Unix(),
	}

	stat := &activity.ActivityStat{ActivityId: int64(a.Id)}
	row := struct {
		Participants int64 `db:"participants"`
		Winners      int64 `db:"winners"`
		StockLeft    int64 `db:"stock_left"`
	}{}
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT IFNULL(participants,0) AS participants, IFNULL(winners,0) AS winners, IFNULL(stock_left,0) AS stock_left FROM `activity_stat` WHERE activity_id=? LIMIT 1",
		a.Id); err == nil {
		stat.Participants = row.Participants
		stat.Winners = row.Winners
		stat.StockLeft = row.StockLeft
	}

	// Token issuance is the caller's responsibility (mall-api), since the
	// authenticated user_id only exists at the API gateway boundary. Activity
	// detail leaves ParticipationToken empty here; mall-api fills it for
	// token-required types (lottery/seckill).
	return &activity.ActivityDetail{Activity: info, Stat: stat}, nil
}
