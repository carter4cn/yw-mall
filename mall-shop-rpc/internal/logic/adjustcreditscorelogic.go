package logic

import (
	"context"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdjustCreditScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdjustCreditScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustCreditScoreLogic {
	return &AdjustCreditScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdjustCreditScoreLogic) AdjustCreditScore(in *shop.AdjustCreditScoreReq) (*shop.OkResp, error) {
	now := time.Now().Unix()

	// Read current → compute new locally → write. Computing the new score
	// in-process means the F-5 thresholds operate on the value we just
	// wrote, even when ProxySQL routes the post-write SELECT to a replica
	// that lags briefly behind the master.
	var current int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &current,
		"SELECT credit_score FROM shop WHERE id = ?", in.ShopId); err != nil {
		return nil, err
	}
	next := current + int64(in.Delta)
	if next < 0 {
		next = 0
	}
	if next > 200 {
		next = 200
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE shop SET credit_score = ?, update_time = ? WHERE id = ?",
		next, now, in.ShopId); err != nil {
		return nil, err
	}
	l.Logger.Infof("shop %d credit %d → %d (operator=%d) reason=%s",
		in.ShopId, current, next, in.OperatorId, in.Reason)

	l.applyThresholdRules(in.ShopId, in.OperatorId, next)

	return &shop.OkResp{Ok: true}, nil
}

// applyThresholdRules enforces the F-5 auto-restriction bands:
//
//	score < 30      → ban shop + ban_publish + ban_promote
//	30 ≤ score < 50 → ban_publish
//	50 ≤ score < 70 → warning log only
//	score ≥ 70      → no action (recovery is manual)
//
// Restrictions are append-only; admin must explicitly lift them as a shop
// recovers (matches PRD §3.3: "操作需填写理由并记录日志, 不可删除供审计").
func (l *AdjustCreditScoreLogic) applyThresholdRules(shopID, operatorID, score int64) {
	switch {
	case score < 30:
		l.banShop(shopID, operatorID, score)
		l.addRestriction(shopID, "ban_publish", "auto: credit_score<30", operatorID)
		l.addRestriction(shopID, "ban_promote", "auto: credit_score<30", operatorID)
	case score < 50:
		l.addRestriction(shopID, "ban_publish", "auto: credit_score<50", operatorID)
	case score < 70:
		l.Logger.Infof("F-5 warn: shop %d credit_score=%d (warning band)", shopID, score)
	}
}

func (l *AdjustCreditScoreLogic) banShop(shopID, operatorID, score int64) {
	now := time.Now().Unix()
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE shop SET status = 2, update_time = ? WHERE id = ? AND status != 2",
		now, shopID); err != nil {
		l.Logger.Errorf("F-5 ban shop %d: %v", shopID, err)
		return
	}
	l.Logger.Infow("F-5 auto-ban shop",
		logx.Field("shop_id", shopID),
		logx.Field("score", score),
		logx.Field("operator_id", operatorID),
	)
}

func (l *AdjustCreditScoreLogic) addRestriction(shopID int64, restriction, reason string, operatorID int64) {
	if l.svcCtx.RiskDB == nil {
		l.Logger.Errorf("F-5 RiskDB not configured; cannot persist restriction %q for shop %d", restriction, shopID)
		return
	}
	now := time.Now().Unix()
	var existing int64
	err := l.svcCtx.RiskDB.QueryRowCtx(l.ctx, &existing,
		"SELECT id FROM shop_restriction WHERE shop_id = ? AND restriction = ? "+
			"AND (expire_time = 0 OR expire_time > ?) LIMIT 1",
		shopID, restriction, now)
	if err == nil && existing > 0 {
		return
	}
	if _, err := l.svcCtx.RiskDB.ExecCtx(l.ctx,
		"INSERT INTO shop_restriction (shop_id, restriction, reason, operator_id, expire_time, create_time) "+
			"VALUES (?, ?, ?, ?, 0, ?)",
		shopID, restriction, reason, operatorID, now); err != nil {
		l.Logger.Errorf("F-5 add restriction %q on shop %d: %v", restriction, shopID, err)
		return
	}
	l.Logger.Infow("F-5 auto-restriction applied",
		logx.Field("shop_id", shopID),
		logx.Field("restriction", restriction),
		logx.Field("reason", reason),
	)
}
