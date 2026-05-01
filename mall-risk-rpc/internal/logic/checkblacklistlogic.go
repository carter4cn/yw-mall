package logic

import (
	"context"
	"fmt"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckBlacklistLogic {
	return &CheckBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckBlacklist hits the per-(subject_type) Redis set for a fast yes/no, then
// confirms the hit against the blacklist table to weed out expired entries.
// Stand-in for a real Redis-Bloom filter — the set works for our scale and
// avoids depending on an extension module.
func (l *CheckBlacklistLogic) CheckBlacklist(in *risk.CheckBlacklistReq) (*risk.CheckBlacklistResp, error) {
	setKey := fmt.Sprintf("risk:bl:%s", in.SubjectType)
	hit, err := l.svcCtx.Redis.SismemberCtx(l.ctx, setKey, in.SubjectValue)
	if err != nil {
		return nil, err
	}
	if !hit {
		return &risk.CheckBlacklistResp{Blacklisted: false}, nil
	}
	// confirm against DB so an expired blacklist entry doesn't keep blocking
	row := struct {
		Reason    string `db:"reason"`
		ExpiresAt int64  `db:"expires_at"`
	}{}
	q := "SELECT reason, IFNULL(UNIX_TIMESTAMP(expires_at),0) AS expires_at FROM `blacklist` WHERE subject_type=? AND subject_value=? ORDER BY id DESC LIMIT 1 FOR UPDATE"
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row, q, in.SubjectType, in.SubjectValue); err != nil {
		// row gone but Redis still has it — clean up and return not blacklisted
		_, _ = l.svcCtx.Redis.SremCtx(l.ctx, setKey, in.SubjectValue)
		return &risk.CheckBlacklistResp{Blacklisted: false}, nil
	}
	if row.ExpiresAt > 0 && row.ExpiresAt < time.Now().Unix() {
		_, _ = l.svcCtx.Redis.SremCtx(l.ctx, setKey, in.SubjectValue)
		return &risk.CheckBlacklistResp{Blacklisted: false}, nil
	}
	return &risk.CheckBlacklistResp{Blacklisted: true, Reason: row.Reason}, nil
}
