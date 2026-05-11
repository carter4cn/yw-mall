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

type ReviewShopApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewShopApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewShopApplicationLogic {
	return &ReviewShopApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReviewShopApplicationLogic) ReviewShopApplication(in *shop.ReviewShopApplicationReq) (*shop.OkResp, error) {
	// action: 1=approve, 2=reject, 3=need_more_info
	var newStatus int32
	switch in.Action {
	case 1:
		newStatus = 1
	case 2:
		newStatus = 2
	case 3:
		newStatus = 3
	default:
		return nil, errors.New("invalid action")
	}

	now := time.Now().Unix()
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		var app applicationRow
		if e := sess.QueryRowCtx(ctx, &app,
			"SELECT "+applicationCols+" FROM shop_application WHERE id=? LIMIT 1 FOR UPDATE", in.ApplicationId); e != nil {
			if e == sqlx.ErrNotFound {
				return errors.New("application not found")
			}
			return e
		}
		if app.Status != 0 && app.Status != 3 {
			return errors.New("application already reviewed")
		}

		var shopId int64
		if newStatus == 1 {
			res, e := sess.ExecCtx(ctx,
				`INSERT INTO shop (name, logo, banner, description, rating, product_count, follow_count, status, create_time, update_time, owner_user_id, credit_score, level, contact_phone, business_license)
				 VALUES (?, ?, '', ?, 5.00, 0, 0, 1, ?, ?, ?, 100, 1, ?, ?)`,
				app.ShopName, app.Logo, app.Description, now, now, app.UserId, app.ContactPhone, app.BusinessLicense)
			if e != nil {
				return e
			}
			shopId, _ = res.LastInsertId()
		}

		if _, e := sess.ExecCtx(ctx,
			"UPDATE shop_application SET status=?, review_remark=?, reviewer_id=?, shop_id=?, update_time=? WHERE id=?",
			newStatus, in.Remark, in.ReviewerId, shopId, now, in.ApplicationId); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
