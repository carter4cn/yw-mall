package logic

import (
	"context"

	"mall-activity-rpc/activity"
	"mall-api/internal/svc"
	"mall-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivityListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityListLogic {
	return &ActivityListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityListLogic) ActivityList(req *types.ActivityListReq) (*types.ActivityListResp, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.PageSize
	if size <= 0 {
		size = 20
	}
	res, err := l.svcCtx.ActivityRpc.ListActivities(l.ctx, &activity.ListActivitiesReq{
		Type:     req.Type,
		Status:   req.Status,
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return nil, err
	}
	out := &types.ActivityListResp{
		Activities: make([]types.ActivityItem, 0, len(res.Activities)),
		Total:      res.Total,
	}
	for _, a := range res.Activities {
		out.Activities = append(out.Activities, types.ActivityItem{
			Id:          a.Id,
			Code:        a.Code,
			Title:       a.Title,
			Type:        a.Type,
			Status:      a.Status,
			StartTime:   a.StartTime,
			EndTime:     a.EndTime,
			Description: a.Description,
		})
	}
	return out, nil
}
