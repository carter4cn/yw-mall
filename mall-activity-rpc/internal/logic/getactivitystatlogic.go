package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityStatLogic {
	return &GetActivityStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityStatLogic) GetActivityStat(in *activity.IdReq) (*activity.ActivityStat, error) {
	row := struct {
		Participants int64 `db:"participants"`
		Winners      int64 `db:"winners"`
		StockLeft    int64 `db:"stock_left"`
	}{}
	stat := &activity.ActivityStat{ActivityId: in.Id}
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT IFNULL(participants,0) AS participants, IFNULL(winners,0) AS winners, IFNULL(stock_left,0) AS stock_left FROM `activity_stat` WHERE activity_id=? LIMIT 1",
		in.Id); err == nil {
		stat.Participants = row.Participants
		stat.Winners = row.Winners
		stat.StockLeft = row.StockLeft
	}
	return stat, nil
}
