package logic

import (
	"context"
	"database/sql"
	"errors"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRefundRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRefundRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRefundRequestLogic {
	return &GetRefundRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRefundRequestLogic) GetRefundRequest(in *order.GetRefundRequestReq) (*order.RefundRequest, error) {
	row, err := loadRefundById(l.ctx, l.svcCtx, in.Id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("refund request not found")
	}
	if err != nil {
		return nil, err
	}
	if in.UserId > 0 && row.UserId != in.UserId {
		return nil, errors.New("refund request not found")
	}
	return toRefundProto(row), nil
}
