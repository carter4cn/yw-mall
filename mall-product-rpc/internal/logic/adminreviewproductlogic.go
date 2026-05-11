package logic

import (
	"context"
	"errors"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminReviewProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminReviewProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminReviewProductLogic {
	return &AdminReviewProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminReviewProductLogic) AdminReviewProduct(in *product.AdminReviewProductReq) (*product.OkResp, error) {
	var newStatus int32
	switch in.Action {
	case 1:
		newStatus = 1
	case 2:
		newStatus = 2
	default:
		return nil, errors.New("invalid action")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE product SET review_status=?, review_remark=? WHERE id=?",
		newStatus, in.Remark, in.Id); err != nil {
		return nil, err
	}
	l.Logger.Infof("product %d reviewed action=%d reviewer=%d", in.Id, in.Action, in.ReviewerId)
	return &product.OkResp{Ok: true}, nil
}
