// Code scaffolded manually for S1.3 mock-confirm endpoint.

package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmMockPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmMockPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmMockPayLogic {
	return &ConfirmMockPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmMockPayLogic) ConfirmMockPay(req *types.ConfirmMockPayReq) (*types.OkResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	if _, err := l.svcCtx.PaymentRpc.ConfirmMockPay(l.ctx, &payment.ConfirmMockPayReq{
		OrderId: req.OrderId,
		UserId:  userId,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
