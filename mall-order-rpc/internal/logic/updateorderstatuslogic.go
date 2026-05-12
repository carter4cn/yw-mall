package logic

import (
	"context"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateOrderStatusLogic) UpdateOrderStatus(in *order.UpdateOrderStatusReq) (*order.UpdateOrderStatusResp, error) {
	o, err := l.svcCtx.OrderModel.FindOne(l.ctx, uint64(in.Id))
	if err == model.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	o.Status = int64(in.Status)
	if err := l.svcCtx.OrderModel.Update(l.ctx, o); err != nil {
		return nil, err
	}

	// S1.5: stamp the per-transition timeline column. H-2's complete_time
	// branch is preserved for the settlement worker's T+N cooling-off window.
	switch in.Status {
	case 1:
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE `order` SET pay_time = UNIX_TIMESTAMP() WHERE id = ? AND pay_time = 0",
			in.Id,
		)
	case 2:
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE `order` SET ship_time = UNIX_TIMESTAMP() WHERE id = ? AND ship_time = 0",
			in.Id,
		)
	case 3:
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE `order` SET complete_time = UNIX_TIMESTAMP() WHERE id = ? AND complete_time = 0",
			in.Id,
		)
	case 4:
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE `order` SET cancel_time = UNIX_TIMESTAMP(), cancel_reason = ? WHERE id = ? AND cancel_time = 0",
			"manual:update_status", in.Id,
		)
	}

	return &order.UpdateOrderStatusResp{}, nil
}
