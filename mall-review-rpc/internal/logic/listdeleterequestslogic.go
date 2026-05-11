package logic

import (
	"context"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDeleteRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDeleteRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDeleteRequestsLogic {
	return &ListDeleteRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type deleteRequestRow struct {
	Id          int64  `db:"id"`
	ReviewId    int64  `db:"review_id"`
	ShopId      int64  `db:"shop_id"`
	Reason      string `db:"reason"`
	Status      int64  `db:"status"`
	AdminRemark string `db:"admin_remark"`
	AdminId     int64  `db:"admin_id"`
	CreateTime  int64  `db:"create_time"`
}

func (l *ListDeleteRequestsLogic) ListDeleteRequests(in *review.ListDeleteRequestsReq) (*review.ListDeleteRequestsResp, error) {
	where := ""
	args := []any{}
	if in.Status >= 0 {
		where = "WHERE status = ?"
		args = append(args, in.Status)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM review_delete_request "+where, args...); err != nil {
		return nil, err
	}

	_, pageSize, offset := clampPaging(in.Page, in.PageSize)
	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, pageSize, offset)

	rows := []*deleteRequestRow{}
	q := "SELECT id, review_id, shop_id, reason, status, admin_remark, admin_id, create_time FROM review_delete_request " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, pagedArgs...); err != nil {
		return nil, err
	}

	out := make([]*review.ReviewDeleteRequest, 0, len(rows))
	for _, r := range rows {
		out = append(out, &review.ReviewDeleteRequest{
			Id:          r.Id,
			ReviewId:    r.ReviewId,
			ShopId:      r.ShopId,
			Reason:      r.Reason,
			Status:      int32(r.Status),
			AdminRemark: r.AdminRemark,
			AdminId:     r.AdminId,
			CreateTime:  r.CreateTime,
		})
	}
	return &review.ListDeleteRequestsResp{Requests: out, Total: total}, nil
}
