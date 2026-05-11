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

	// H-2: stamp complete_time on transition to status=3 so the settlement
	// worker can apply the T+N cooling-off window.
	if in.Status == 3 {
		_, _ = l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE `order` SET complete_time = UNIX_TIMESTAMP() WHERE id = ? AND complete_time = 0",
			in.Id,
		)
	}

	return &order.UpdateOrderStatusResp{}, nil
}
