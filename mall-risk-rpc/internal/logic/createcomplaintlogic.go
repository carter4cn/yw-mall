package logic

import (
	"context"
	"time"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateComplaintLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateComplaintLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateComplaintLogic {
	return &CreateComplaintLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateComplaintLogic) CreateComplaint(in *risk.CreateComplaintReq) (*risk.CreateComplaintResp, error) {
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO complaint_ticket
		 (complainant_type, complainant_id, defendant_type, defendant_id, order_id,
		  category, content, evidence_urls, status, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		in.ComplainantType, in.ComplainantId, in.DefendantType, in.DefendantId, in.OrderId,
		in.Category, in.Content, in.EvidenceUrls, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &risk.CreateComplaintResp{Id: id}, nil
}
