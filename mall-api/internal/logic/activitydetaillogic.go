package logic

import (
	"context"
	"encoding/json"

	"mall-activity-rpc/activity"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivityDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityDetailLogic {
	return &ActivityDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityDetailLogic) ActivityDetail(req *types.ActivityDetailReq) (*types.ActivityDetailResp, error) {
	res, err := l.svcCtx.ActivityRpc.GetActivity(l.ctx, &activity.IdReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	resp := &types.ActivityDetailResp{
		ParticipationToken: res.ParticipationToken,
	}
	if res.Activity != nil {
		resp.Id = res.Activity.Id
		resp.Code = res.Activity.Code
		resp.Title = res.Activity.Title
		resp.Description = res.Activity.Description
		resp.Type = res.Activity.Type
		resp.Status = res.Activity.Status
		resp.StartTime = res.Activity.StartTime
		resp.EndTime = res.Activity.EndTime
	}
	if res.Stat != nil {
		resp.Participants = res.Stat.Participants
		resp.Winners = res.Stat.Winners
		resp.StockLeft = res.Stat.StockLeft
	}

	// Issue a short-lived HMAC token for activity types that require one. The
	// token binds (user, activity, exp) so it can't be sideloaded across users.
	// Done here in mall-api because the authenticated uid only exists at the
	// gateway boundary — activity-rpc has no JWT context.
	if needsToken(resp.Type) {
		if uid := uidFromCtx(l.ctx); uid > 0 {
			tok, err := l.svcCtx.RiskRpc.IssueToken(l.ctx, &risk.IssueTokenReq{
				UserId:     uid,
				ActivityId: req.Id,
				TtlSeconds: 300,
			})
			if err != nil {
				l.Logger.Errorf("risk.IssueToken activity=%d user=%d: %v", req.Id, uid, err)
			} else {
				resp.ParticipationToken = tok.Token
			}
		}
	}
	return resp, nil
}

func needsToken(activityType string) bool {
	return activityType == "lottery" || activityType == "seckill"
}

func uidFromCtx(ctx context.Context) int64 {
	v, ok := ctx.Value("uid").(json.Number)
	if !ok {
		return 0
	}
	id, _ := v.Int64()
	return id
}
