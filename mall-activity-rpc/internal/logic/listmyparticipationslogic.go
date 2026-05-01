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

type ListMyParticipationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyParticipationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyParticipationsLogic {
	return &ListMyParticipationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type myPartRow struct {
	Id                 int64     `db:"id"`
	ActivityId         int64     `db:"activity_id"`
	UserId             int64     `db:"user_id"`
	Sequence           int64     `db:"sequence"`
	WorkflowInstanceId int64     `db:"workflow_instance_id"`
	Status             string    `db:"status"`
	PayloadJson        string    `db:"payload_json"`
	CreateTime         time.Time `db:"create_time"`
}

func (l *ListMyParticipationsLogic) ListMyParticipations(in *activity.ListMyParticipationsReq) (*activity.ListMyParticipationsResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	clauses := []string{"user_id=?"}
	args := []any{in.UserId}
	if in.ActivityId > 0 {
		clauses = append(clauses, "activity_id=?")
		args = append(args, in.ActivityId)
	}
	where := strings.Join(clauses, " AND ")

	rows := []myPartRow{}
	q := fmt.Sprintf("SELECT id, activity_id, user_id, sequence, workflow_instance_id, status, IFNULL(payload_json,'') AS payload_json, create_time FROM `participation_record` WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d", where, size, offset)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM `participation_record` WHERE %s", where), args...); err != nil {
		return nil, err
	}
	out := make([]*activity.ParticipationRecord, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &activity.ParticipationRecord{
			Id:                 r.Id,
			ActivityId:         r.ActivityId,
			UserId:             r.UserId,
			Sequence:           r.Sequence,
			WorkflowInstanceId: r.WorkflowInstanceId,
			Status:             r.Status,
			PayloadJson:        r.PayloadJson,
			CreateTime:         r.CreateTime.Unix(),
		})
	}
	return &activity.ListMyParticipationsResp{Records: out, Total: total}, nil
}
