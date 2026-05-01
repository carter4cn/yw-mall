package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyRewardsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyRewardsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRewardsLogic {
	return &MyRewardsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyRewardsLogic) MyRewards(req *types.MyRewardsReq) (*types.MyRewardsResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.PageSize
	if size <= 0 {
		size = 20
	}

	res, err := l.svcCtx.RewardRpc.GetMyRewards(l.ctx, &reward.GetMyRewardsReq{
		UserId:   userId,
		Type:     req.Type,
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return nil, err
	}
	out := &types.MyRewardsResp{
		Rewards: make([]types.MyRewardItem, 0, len(res.Records)),
		Total:   res.Total,
	}
	for _, r := range res.Records {
		out.Rewards = append(out.Rewards, types.MyRewardItem{
			Id:          r.Id,
			ActivityId:  r.ActivityId,
			TemplateId:  r.TemplateId,
			Type:        r.Type,
			PayloadJson: r.PayloadJson,
			Status:      r.Status,
			CreateTime:  r.CreateTime,
		})
	}
	return out, nil
}
