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

type ReviewLevelApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewLevelApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLevelApplicationLogic {
	return &ReviewLevelApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReviewLevelApplicationLogic) ReviewLevelApplication(in *shop.ReviewLevelApplicationReq) (*shop.OkResp, error) {
	// action: 1=approve, 2=reject
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
		var app levelApplicationRow
		if e := sess.QueryRowCtx(ctx, &app,
			"SELECT "+levelApplicationCols+" FROM shop_level_application WHERE id=? LIMIT 1 FOR UPDATE", in.ApplicationId); e != nil {
			if e == sqlx.ErrNotFound {
				return errors.New("application not found")
			}
			return e
		}
		if app.Status != 0 {
			return errors.New("application already reviewed")
		}

		if _, e := sess.ExecCtx(ctx,
			"UPDATE shop_level_application SET status=?, admin_id=?, admin_remark=?, update_time=? WHERE id=?",
			newStatus, in.AdminId, in.Remark, now, in.ApplicationId); e != nil {
			return e
		}

		if newStatus == 1 {
			if _, e := sess.ExecCtx(ctx,
				"UPDATE shop SET level=?, update_time=? WHERE id=?",
				app.TargetLevel, now, app.ShopId); e != nil {
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
