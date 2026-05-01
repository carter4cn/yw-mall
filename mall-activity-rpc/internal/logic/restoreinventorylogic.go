package logic

import (
	"context"
	"fmt"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRestoreInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreInventoryLogic {
	return &RestoreInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RestoreInventory is the SAGA compensation for DecrInventory. Adds the
// quantity back to current_stock. Uses idempotency_key to skip duplicate
// retries from DTM.
func (l *RestoreInventoryLogic) RestoreInventory(in *activity.RestoreInventoryReq) (*activity.Empty, error) {
	idemKey := fmt.Sprintf("idem:restore_inv:%s", in.IdempotencyKey)
	if in.IdempotencyKey != "" {
		ok, _ := l.svcCtx.Redis.SetnxExCtx(l.ctx, idemKey, "1", 86400)
		if !ok {
			return &activity.Empty{}, nil
		}
	}
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `activity_inventory_snapshot` SET current_stock = current_stock + ? WHERE activity_id=?",
		in.Quantity, in.ActivityId,
	)
	if err != nil {
		return nil, err
	}
	return &activity.Empty{}, nil
}
