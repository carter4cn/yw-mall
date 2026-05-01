package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddToBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddToBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddToBlacklistLogic {
	return &AddToBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddToBlacklistLogic) AddToBlacklist(in *risk.AddToBlacklistReq) (*risk.Empty, error) {
	var expires sql.NullTime
	if in.ExpiresAt > 0 {
		expires = sql.NullTime{Time: time.Unix(in.ExpiresAt, 0), Valid: true}
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO `blacklist`(subject_type, subject_value, reason, expires_at) VALUES (?,?,?,?)",
		in.SubjectType, in.SubjectValue, in.Reason, expires,
	); err != nil {
		return nil, err
	}
	setKey := fmt.Sprintf("risk:bl:%s", in.SubjectType)
	if _, err := l.svcCtx.Redis.SaddCtx(l.ctx, setKey, in.SubjectValue); err != nil {
		l.Logger.Errorf("blacklist redis SADD %s: %v", setKey, err)
	}
	return &risk.Empty{}, nil
}
