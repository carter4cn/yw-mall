// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"mall-api/internal/svc"
)

type Kuaidi100WebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKuaidi100WebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Kuaidi100WebhookLogic {
	return &Kuaidi100WebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Kuaidi100WebhookLogic) Kuaidi100Webhook() (resp string, err error) {
	// todo: add your logic here and delete this line

	return
}
