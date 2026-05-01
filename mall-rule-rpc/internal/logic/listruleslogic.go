package logic

import (
	"context"
	"fmt"

	"mall-rule-rpc/internal/model"
	"mall-rule-rpc/internal/svc"
	"mall-rule-rpc/rule"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRulesLogic {
	return &ListRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRulesLogic) ListRules(in *rule.ListRulesReq) (*rule.ListRulesResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	rows := make([]model.Rule, 0)
	q := fmt.Sprintf("SELECT `id`, `code`, `description`, `expression`, `lang`, `version`, `status`, `json_schema`, `create_time`, `update_time` FROM `rule` ORDER BY id DESC LIMIT %d OFFSET %d", size, offset)
	if err := rawQuery(l.ctx, l.svcCtx, q, &rows); err != nil {
		return nil, err
	}
	var total int64
	if err := rawCount(l.ctx, l.svcCtx, "SELECT COUNT(*) FROM `rule`", &total); err != nil {
		return nil, err
	}
	out := make([]*rule.RuleInfo, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, &rule.RuleInfo{
			Id:          int64(r.Id),
			Code:        r.Code,
			Description: r.Description,
			Expression:  r.Expression,
			Lang:        r.Lang,
			Version:     int32(r.Version),
			Status:      r.Status,
			JsonSchema:  r.JsonSchema.String,
			CreateTime:  r.CreateTime.Unix(),
			UpdateTime:  r.UpdateTime.Unix(),
		})
	}
	return &rule.ListRulesResp{Rules: out, Total: total}, nil
}
