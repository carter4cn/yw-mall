package logic

import (
	"context"
	"errors"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReviewShopLifecycleRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewShopLifecycleRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewShopLifecycleRequestLogic {
	return &ReviewShopLifecycleRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReviewShopLifecycleRequest approves or rejects a lifecycle change. On approve,
// the shop's status is updated: deactivate=>4, pause=>3, resume=>1.
func (l *ReviewShopLifecycleRequestLogic) ReviewShopLifecycleRequest(in *shop.ReviewShopLifecycleRequestReq) (*shop.OkResp, error) {
	var newStatus int32
	switch in.Action {
	case 1:
		newStatus = 1
	case 2:
		newStatus = 2
	default:
		return nil, errors.New("invalid action")
	}

	now := time.Now().Unix()
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		var req lifecycleRequestRow
		if e := sess.QueryRowCtx(ctx, &req,
			"SELECT "+lifecycleRequestCols+" FROM shop_lifecycle_request WHERE id=? LIMIT 1 FOR UPDATE", in.RequestId); e != nil {
			if e == sqlx.ErrNotFound {
				return errors.New("request not found")
			}
			return e
		}
		if req.Status != 0 {
			return errors.New("request already reviewed")
		}

		if _, e := sess.ExecCtx(ctx,
			"UPDATE shop_lifecycle_request SET status=?, admin_id=?, admin_remark=?, update_time=? WHERE id=?",
			newStatus, in.AdminId, in.Remark, now, in.RequestId); e != nil {
			return e
		}

		if newStatus == 1 {
			var shopStatus int32
			switch req.Action {
			case "deactivate":
				shopStatus = 4
			case "pause":
				shopStatus = 3
			case "resume":
				shopStatus = 1
			default:
				return errors.New("invalid lifecycle action")
			}
			if _, e := sess.ExecCtx(ctx,
				"UPDATE shop SET status=?, update_time=? WHERE id=?",
				shopStatus, now, req.ShopId); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
