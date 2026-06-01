package logic

import (
	"context"
	"database/sql"
	"errors"

	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

// GetActivity 加载活动 + 关联的所有 targets + actions。
func GetActivity(ctx context.Context, svcCtx *svc.ServiceContext, in *promotion.GetActivityReq) (*promotion.GetActivityResp, error) {
	act, err := loadActivityRow(ctx, svcCtx, in.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("activity not found")
		}
		return nil, err
	}
	targets, err := loadActivityTargets(ctx, svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	actions, err := loadActivityActions(ctx, svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	rule, err := loadActivityRule(ctx, svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	act.Targets = targets
	act.Actions = actions
	act.Rule = rule
	return &promotion.GetActivityResp{Activity: act}, nil
}
