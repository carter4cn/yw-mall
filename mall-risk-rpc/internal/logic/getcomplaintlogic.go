package logic

import (
	"context"
	"database/sql"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComplaintLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetComplaintLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComplaintLogic {
	return &GetComplaintLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type complaintRow struct {
	Id              int64          `db:"id"`
	ComplainantType string         `db:"complainant_type"`
	ComplainantId   int64          `db:"complainant_id"`
	DefendantType   string         `db:"defendant_type"`
	DefendantId     int64          `db:"defendant_id"`
	OrderId         int64          `db:"order_id"`
	Category        string         `db:"category"`
	Content         string         `db:"content"`
	EvidenceUrls    sql.NullString `db:"evidence_urls"`
	Status          int64          `db:"status"`
	AdminId         int64          `db:"admin_id"`
	AdminRemark     string         `db:"admin_remark"`
	CreateTime      int64          `db:"create_time"`
	UpdateTime      int64          `db:"update_time"`
}

const complaintCols = "id, complainant_type, complainant_id, defendant_type, defendant_id, order_id, category, content, evidence_urls, status, admin_id, admin_remark, create_time, update_time"

func (l *GetComplaintLogic) GetComplaint(in *risk.GetComplaintReq) (*risk.ComplaintTicket, error) {
	var row complaintRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT "+complaintCols+" FROM complaint_ticket WHERE id = ?", in.Id); err != nil {
		return nil, err
	}
	return rowToComplaint(&row), nil
}

func rowToComplaint(r *complaintRow) *risk.ComplaintTicket {
	return &risk.ComplaintTicket{
		Id:              r.Id,
		ComplainantType: r.ComplainantType,
		ComplainantId:   r.ComplainantId,
		DefendantType:   r.DefendantType,
		DefendantId:     r.DefendantId,
		OrderId:         r.OrderId,
		Category:        r.Category,
		Content:         r.Content,
		EvidenceUrls:    r.EvidenceUrls.String,
		Status:          int32(r.Status),
		AdminId:         r.AdminId,
		AdminRemark:     r.AdminRemark,
		CreateTime:      r.CreateTime,
		UpdateTime:      r.UpdateTime,
	}
}
