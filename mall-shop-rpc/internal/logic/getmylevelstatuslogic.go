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

type GetMyLevelStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyLevelStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyLevelStatusLogic {
	return &GetMyLevelStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type shopLevelInfoRow struct {
	Level       int64   `db:"level"`
	CreditScore int64   `db:"credit_score"`
	Rating      float64 `db:"rating"`
	CreateTime  int64   `db:"create_time"`
}

func (l *GetMyLevelStatusLogic) GetMyLevelStatus(in *shop.GetMyLevelStatusReq) (*shop.MyLevelStatus, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}

	var info shopLevelInfoRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &info,
		"SELECT level, credit_score, rating, create_time FROM shop WHERE id=? LIMIT 1", in.ShopId); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("shop not found")
		}
		return nil, err
	}

	// TODO: aggregate cumulative GMV from mall-order-rpc; placeholder 0 for now
	var currentGmv int64 = 0
	currentMonths := int32(0)
	if info.CreateTime > 0 {
		currentMonths = int32((time.Now().Unix() - info.CreateTime) / 86400 / 30)
	}

	var templates []*levelTemplateRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &templates,
		"SELECT "+levelTemplateCols+" FROM shop_level_template ORDER BY level ASC"); err != nil {
		return nil, err
	}

	var current, next *shop.ShopLevelTemplate
	for _, t := range templates {
		if int32(t.Level) == int32(info.Level) {
			current = toLevelTemplateProto(t)
		}
		if int32(t.Level) == int32(info.Level)+1 {
			next = toLevelTemplateProto(t)
		}
	}

	eligible := false
	if next != nil {
		eligible = currentGmv >= next.MinGmv &&
			int32(info.CreditScore) >= next.MinCreditScore &&
			currentMonths >= next.MinMonths &&
			info.Rating >= next.MinRating
	}

	var pending int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &pending,
		"SELECT COUNT(*) FROM shop_level_application WHERE shop_id=? AND status=0", in.ShopId); err != nil {
		return nil, err
	}

	return &shop.MyLevelStatus{
		CurrentLevel:          int32(info.Level),
		CurrentTemplate:       current,
		NextTemplate:          next,
		CurrentGmv:            currentGmv,
		CurrentCreditScore:    int32(info.CreditScore),
		CurrentMonths:         currentMonths,
		CurrentRating:         info.Rating,
		EligibleForNext:       eligible,
		HasPendingApplication: pending > 0,
	}, nil
}
