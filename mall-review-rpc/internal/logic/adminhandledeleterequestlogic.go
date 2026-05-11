package logic

import (
	"context"
	"errors"
	"time"

	"mall-review-rpc/internal/svc"
	"mall-review-rpc/review"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminHandleDeleteRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminHandleDeleteRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminHandleDeleteRequestLogic {
	return &AdminHandleDeleteRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AdminHandleDeleteRequest approves (action=1) or rejects (action=2) a
// merchant's review takedown request. Approve also soft-deletes the review
// in the same transaction (status=1=admin_soft_deleted per existing schema).
func (l *AdminHandleDeleteRequestLogic) AdminHandleDeleteRequest(in *review.AdminHandleDeleteRequestReq) (*review.OkResp, error) {
	if in.Action != 1 && in.Action != 2 {
		return nil, errors.New("invalid action")
	}
	now := time.Now().Unix()
	newStatus := int32(2) // rejected
	if in.Action == 1 {
		newStatus = 1 // approved
	}

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(_ context.Context, tx sqlx.Session) error {
		// Look up the request to get review_id for the optional soft delete.
		var reviewId int64
		if err := tx.QueryRowCtx(l.ctx, &reviewId,
			"SELECT review_id FROM review_delete_request WHERE id = ? FOR UPDATE", in.RequestId); err != nil {
			return err
		}
		if _, err := tx.ExecCtx(l.ctx,
			"UPDATE review_delete_request SET status=?, admin_remark=?, admin_id=?, update_time=? WHERE id=?",
			newStatus, in.Remark, in.AdminId, now, in.RequestId); err != nil {
			return err
		}
		if in.Action == 1 {
			if _, err := tx.ExecCtx(l.ctx,
				"UPDATE review SET status=1, update_time=NOW() WHERE id=?", reviewId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &review.OkResp{Ok: true}, nil
}
