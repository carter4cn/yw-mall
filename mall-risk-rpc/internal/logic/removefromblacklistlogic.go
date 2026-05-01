package logic

import (
	"context"
	"fmt"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveFromBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveFromBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveFromBlacklistLogic {
	return &RemoveFromBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveFromBlacklistLogic) RemoveFromBlacklist(in *risk.RemoveFromBlacklistReq) (*risk.Empty, error) {
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"DELETE FROM `blacklist` WHERE subject_type=? AND subject_value=?",
		in.SubjectType, in.SubjectValue,
	); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.Redis.SremCtx(l.ctx, fmt.Sprintf("risk:bl:%s", in.SubjectType), in.SubjectValue); err != nil {
		l.Logger.Errorf("blacklist redis SREM: %v", err)
	}
	return &risk.Empty{}, nil
}
