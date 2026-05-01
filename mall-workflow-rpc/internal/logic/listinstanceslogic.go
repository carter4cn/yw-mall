package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListInstancesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListInstancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInstancesLogic {
	return &ListInstancesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type instRow struct {
	Id           int64     `db:"id"`
	DefinitionId int64     `db:"definition_id"`
	ActivityId   int64     `db:"activity_id"`
	UserId       int64     `db:"user_id"`
	State        string    `db:"state"`
	PayloadJson  string    `db:"payload_json"`
	Version      int32     `db:"version"`
	LastEventAt  time.Time `db:"last_event_at"`
	CreateTime   time.Time `db:"create_time"`
}

func (l *ListInstancesLogic) ListInstances(in *workflow.ListInstancesReq) (*workflow.ListInstancesResp, error) {
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
	if in.ActivityId > 0 {
		clauses = append(clauses, "activity_id=?")
		args = append(args, in.ActivityId)
	}
	if in.UserId > 0 {
		clauses = append(clauses, "user_id=?")
		args = append(args, in.UserId)
	}
	if in.State != "" {
		clauses = append(clauses, "state=?")
		args = append(args, in.State)
	}
	where := strings.Join(clauses, " AND ")

	rows := []instRow{}
	q := fmt.Sprintf("SELECT id, definition_id, activity_id, user_id, state, IFNULL(payload_json,'') AS payload_json, version, last_event_at, create_time FROM `workflow_instance` WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d", where, size, offset)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	var total int64
	cq := fmt.Sprintf("SELECT COUNT(*) FROM `workflow_instance` WHERE %s", where)
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, cq, args...); err != nil {
		return nil, err
	}
	out := make([]*workflow.WorkflowInstance, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &workflow.WorkflowInstance{
			Id:           r.Id,
			DefinitionId: r.DefinitionId,
			ActivityId:   r.ActivityId,
			UserId:       r.UserId,
			State:        r.State,
			PayloadJson:  r.PayloadJson,
			Version:      r.Version,
			LastEventAt:  r.LastEventAt.Unix(),
			CreateTime:   r.CreateTime.Unix(),
		})
	}
	return &workflow.ListInstancesResp{Instances: out, Total: total}, nil
}
