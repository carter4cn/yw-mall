package logic

import (
	"context"
	"errors"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSensitiveWordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSensitiveWordLogic {
	return &CreateSensitiveWordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSensitiveWordLogic) CreateSensitiveWord(in *risk.CreateSensitiveWordReq) (*risk.CreateSensitiveWordResp, error) {
	if in.Word == "" {
		return nil, errors.New("word required")
	}
	category := in.Category
	if category == "" {
		category = "general"
	}
	action := in.Action
	if action == "" {
		action = "flag"
	}
	if action != "flag" && action != "block" {
		return nil, errors.New("action must be flag or block")
	}
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO sensitive_word (word, category, action, status, create_time)
		 VALUES (?, ?, ?, 1, ?)
		 ON DUPLICATE KEY UPDATE category=VALUES(category), action=VALUES(action), status=1`,
		in.Word, category, action, now)
	if err != nil {
		return nil, err
	}
	sensitiveCache.invalidate()
	id, _ := res.LastInsertId()
	return &risk.CreateSensitiveWordResp{Id: id}, nil
}
