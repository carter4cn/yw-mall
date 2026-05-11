package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SubmitLevelApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitLevelApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitLevelApplicationLogic {
	return &SubmitLevelApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitLevelApplicationLogic) SubmitLevelApplication(in *shop.SubmitLevelApplicationReq) (*shop.SubmitLevelApplicationResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	if in.TargetLevel <= 0 {
		return nil, errors.New("target_level required")
	}

	// Reject if a pending application already exists.
	var pending int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &pending,
		"SELECT COUNT(*) FROM shop_level_application WHERE shop_id=? AND status=0", in.ShopId); err != nil {
		return nil, err
	}
	if pending > 0 {
		return nil, errors.New("pending application exists")
	}

	var info shopLevelInfoRow
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &info,
		"SELECT level, credit_score, rating, create_time FROM shop WHERE id=? LIMIT 1", in.ShopId); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errors.New("shop not found")
		}
		return nil, err
	}

	currentMonths := int32(0)
	if info.CreateTime > 0 {
		currentMonths = int32((time.Now().Unix() - info.CreateTime) / 86400 / 30)
	}

	snapshot, _ := json.Marshal(map[string]any{
		"gmv":          0, // TODO aggregate from mall-order-rpc
		"credit_score": info.CreditScore,
		"months":       currentMonths,
		"rating":       info.Rating,
	})

	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO shop_level_application (shop_id, current_level, target_level, snapshot, status, admin_id, admin_remark, create_time, update_time)
		 VALUES (?, ?, ?, ?, 0, 0, '', ?, ?)`,
		in.ShopId, info.Level, in.TargetLevel, string(snapshot), now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &shop.SubmitLevelApplicationResp{ApplicationId: id}, nil
}
