package logic

import (
	"context"
	"database/sql"

	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTemplatesLogic {
	return &ListTemplatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type templateRow struct {
	Id                uint64         `db:"id"`
	Code              string         `db:"code"`
	Type              string         `db:"type"`
	PayloadSchemaJson sql.NullString `db:"payload_schema_json"`
	MaxValue          int64          `db:"max_value"`
	Status            string         `db:"status"`
	Description       string         `db:"description"`
}

func (l *ListTemplatesLogic) ListTemplates(in *reward.ListTemplatesReq) (*reward.ListTemplatesResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	args := []any{}
	where := "1=1"
	if in.Type != "" {
		where += " AND `type` = ?"
		args = append(args, in.Type)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM `reward_template` WHERE "+where, args...); err != nil {
		return nil, err
	}

	var rows []templateRow
	q := "SELECT id, code, `type`, payload_schema_json, max_value, status, description FROM `reward_template` WHERE " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, append(args, size, offset)...); err != nil {
		return nil, err
	}

	out := &reward.ListTemplatesResp{Total: total, Templates: make([]*reward.RewardTemplate, 0, len(rows))}
	for _, r := range rows {
		out.Templates = append(out.Templates, &reward.RewardTemplate{
			Id:                int64(r.Id),
			Code:              r.Code,
			Type:              r.Type,
			PayloadSchemaJson: r.PayloadSchemaJson.String,
			MaxValue:          r.MaxValue,
			Status:            r.Status,
			Description:       r.Description,
		})
	}
	return out, nil
}
