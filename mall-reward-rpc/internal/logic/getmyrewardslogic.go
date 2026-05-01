package logic

import (
	"context"
	"database/sql"
	"time"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRewardsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyRewardsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRewardsLogic {
	return &GetMyRewardsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type myRewardRow struct {
	Id                 uint64         `db:"id"`
	UserId             int64          `db:"user_id"`
	ActivityId         int64          `db:"activity_id"`
	WorkflowInstanceId int64          `db:"workflow_instance_id"`
	TemplateId         int64          `db:"template_id"`
	Type               string         `db:"type"`
	PayloadJson        sql.NullString `db:"payload_json"`
	Status             string         `db:"status"`
	IdempotencyKey     string         `db:"idempotency_key"`
	Version            int64          `db:"version"`
	CreateTime         time.Time      `db:"create_time"`
	UpdateTime         time.Time      `db:"update_time"`
}

func (l *GetMyRewardsLogic) GetMyRewards(in *reward.GetMyRewardsReq) (*reward.GetMyRewardsResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	where := "user_id = ?"
	args := []any{in.UserId}
	if in.Type != "" {
		where += " AND `type` = ?"
		args = append(args, in.Type)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM `reward_record` WHERE "+where, args...); err != nil {
		return nil, err
	}

	var rows []myRewardRow
	q := "SELECT id, user_id, activity_id, workflow_instance_id, template_id, `type`, payload_json, status, idempotency_key, version, create_time, update_time FROM `reward_record` WHERE " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, append(args, size, offset)...); err != nil {
		return nil, err
	}

	out := &reward.GetMyRewardsResp{Total: total, Records: make([]*reward.RewardRecord, 0, len(rows))}
	for _, r := range rows {
		out.Records = append(out.Records, &reward.RewardRecord{
			Id:                 int64(r.Id),
			UserId:             r.UserId,
			ActivityId:         r.ActivityId,
			WorkflowInstanceId: r.WorkflowInstanceId,
			TemplateId:         r.TemplateId,
			Type:               r.Type,
			PayloadJson:        r.PayloadJson.String,
			Status:             r.Status,
			IdempotencyKey:     r.IdempotencyKey,
			Version:            int32(r.Version),
			CreateTime:         r.CreateTime.Unix(),
			UpdateTime:         r.UpdateTime.Unix(),
		})
	}
	return out, nil
}
