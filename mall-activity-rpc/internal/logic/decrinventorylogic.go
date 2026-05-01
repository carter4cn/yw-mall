package logic

import (
	"context"
	"fmt"
	"strconv"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DecrInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDecrInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DecrInventoryLogic {
	return &DecrInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DecrInventory is the server-internal entry called by SAGA branches when
// they need to deduct inventory under DTM-controlled atomicity. The Lua
// hot path in Participate already touches Redis; this method also decrements
// the persisted snapshot in MySQL so reconciliation has a paper trail.
func (l *DecrInventoryLogic) DecrInventory(in *activity.DecrInventoryReq) (*activity.DecrInventoryResp, error) {
	// idempotency check via Redis SET NX
	idemKey := fmt.Sprintf("idem:decr_inv:%s", in.IdempotencyKey)
	if in.IdempotencyKey != "" {
		ok, _ := l.svcCtx.Redis.SetnxExCtx(l.ctx, idemKey, "1", 86400)
		if !ok {
			// already processed; return current stock
			left, _ := l.currentStock(in.ActivityId)
			return &activity.DecrInventoryResp{StockLeft: left, Ok: true}, nil
		}
	}

	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `activity_inventory_snapshot` SET current_stock = current_stock - ? WHERE activity_id=? AND current_stock >= ?",
		in.Quantity, in.ActivityId, in.Quantity,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &activity.DecrInventoryResp{StockLeft: 0, Ok: false}, nil
	}
	left, _ := l.currentStock(in.ActivityId)
	return &activity.DecrInventoryResp{StockLeft: left, Ok: true}, nil
}

func (l *DecrInventoryLogic) currentStock(activityId int64) (int64, error) {
	var n int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &n,
		"SELECT IFNULL(SUM(current_stock),0) FROM `activity_inventory_snapshot` WHERE activity_id=?",
		activityId,
	)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// keep strconv import live (used during JSON-payload helpers in sibling files)
var _ = strconv.Itoa
