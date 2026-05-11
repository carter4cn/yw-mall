package logic

import (
	"context"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAdminStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAdminStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAdminStatusLogic {
	return &UpdateAdminStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAdminStatusLogic) UpdateAdminStatus(in *user.UpdateAdminStatusReq) (*user.OkResp, error) {
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE admin_user SET status=?, update_time=? WHERE id=?",
		in.Status, now, in.Id); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}
