package logic

import (
	"context"
	"errors"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSensitiveWordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSensitiveWordLogic {
	return &DeleteSensitiveWordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteSensitiveWord soft-deletes by setting status=0.
func (l *DeleteSensitiveWordLogic) DeleteSensitiveWord(in *risk.IdReq) (*risk.Empty, error) {
	if in.Id <= 0 {
		return nil, errors.New("id required")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE sensitive_word SET status=0 WHERE id=?", in.Id); err != nil {
		return nil, err
	}
	sensitiveCache.invalidate()
	return &risk.Empty{}, nil
}
