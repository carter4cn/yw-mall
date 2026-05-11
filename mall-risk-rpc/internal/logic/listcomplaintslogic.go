package logic

import (
	"context"
	"strings"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListComplaintsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListComplaintsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListComplaintsLogic {
	return &ListComplaintsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListComplaintsLogic) ListComplaints(in *risk.ListComplaintsReq) (*risk.ListComplaintsResp, error) {
	conds := []string{}
	args := []any{}
	if in.Status >= 0 {
		conds = append(conds, "status = ?")
		args = append(args, in.Status)
	}
	if in.DefendantType != "" {
		conds = append(conds, "defendant_type = ?")
		args = append(args, in.DefendantType)
	}
	if in.DefendantId > 0 {
		conds = append(conds, "defendant_id = ?")
		args = append(args, in.DefendantId)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM complaint_ticket "+where, args...); err != nil {
		return nil, err
	}

	page, pageSize := in.Page, in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, pageSize, offset)

	rows := []*complaintRow{}
	q := "SELECT " + complaintCols + " FROM complaint_ticket " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, pagedArgs...); err != nil {
		return nil, err
	}

	out := make([]*risk.ComplaintTicket, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToComplaint(r))
	}
	return &risk.ListComplaintsResp{Tickets: out, Total: total}, nil
}
