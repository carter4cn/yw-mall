package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRiskScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRiskScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRiskScoreLogic {
	return &GetRiskScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRiskScoreLogic) GetRiskScore(in *risk.GetRiskScoreReq) (*risk.GetRiskScoreResp, error) {
	// todo: add your logic here and delete this line

	return &risk.GetRiskScoreResp{}, nil
}
