package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveShopRestrictionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveShopRestrictionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveShopRestrictionLogic {
	return &RemoveShopRestrictionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveShopRestrictionLogic) RemoveShopRestriction(in *risk.RemoveShopRestrictionReq) (*risk.Empty, error) {
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"DELETE FROM shop_restriction WHERE id = ?", in.Id); err != nil {
		return nil, err
	}
	return &risk.Empty{}, nil
}
