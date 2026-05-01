package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListActivitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListActivitiesLogic {
	return &ListActivitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type listRow struct {
	Id          int64     `db:"id"`
	Code        string    `db:"code"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Type        string    `db:"type"`
	Status      string    `db:"status"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
}

func (l *ListActivitiesLogic) ListActivities(in *activity.ListActivitiesReq) (*activity.ListActivitiesResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size
	clauses := []string{"1=1"}
	args := []any{}
	if in.Type != "" {
		clauses = append(clauses, "type=?")
		args = append(args, in.Type)
	}
	if in.Status != "" {
		clauses = append(clauses, "status=?")
		args = append(args, in.Status)
	}
	where := strings.Join(clauses, " AND ")
	rows := []listRow{}
	q := fmt.Sprintf("SELECT id, code, title, IFNULL(description,'') AS description, type, status, start_time, end_time FROM `activity` WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d", where, size, offset)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM `activity` WHERE %s", where), args...); err != nil {
		return nil, err
	}
	out := make([]*activity.ActivityInfo, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &activity.ActivityInfo{
			Id:          r.Id,
			Code:        r.Code,
			Title:       r.Title,
			Description: r.Description,
			Type:        r.Type,
			Status:      r.Status,
			StartTime:   r.StartTime.Unix(),
			EndTime:     r.EndTime.Unix(),
		})
	}
	return &activity.ListActivitiesResp{Activities: out, Total: total}, nil
}
