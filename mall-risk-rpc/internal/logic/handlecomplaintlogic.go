package logic

import (
	"context"
	"errors"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleComplaintLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleComplaintLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleComplaintLogic {
	return &HandleComplaintLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleComplaint maps action codes to closing statuses:
//   1 → closed_support  (status=2)
//   2 → closed_dismiss  (status=3)
//   3 → closed_mediate  (status=4)
func (l *HandleComplaintLogic) HandleComplaint(in *risk.HandleComplaintReq) (*risk.Empty, error) {
	var status int32
	switch in.Action {
	case 1:
		status = 2
	case 2:
		status = 3
	case 3:
		status = 4
	default:
		return nil, errors.New("invalid action")
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE complaint_ticket SET status=?, admin_id=?, admin_remark=?, update_time=? WHERE id=?",
		status, in.AdminId, in.Remark, now, in.Id); err != nil {
		return nil, err
	}
	return &risk.Empty{}, nil
}
