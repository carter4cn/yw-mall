package logic

import (
	"context"
	"errors"
	"time"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitShopLifecycleRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitShopLifecycleRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitShopLifecycleRequestLogic {
	return &SubmitShopLifecycleRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitShopLifecycleRequestLogic) SubmitShopLifecycleRequest(in *shop.SubmitShopLifecycleRequestReq) (*shop.SubmitShopLifecycleRequestResp, error) {
	if in.ShopId <= 0 {
		return nil, errors.New("shop_id required")
	}
	switch in.Action {
	case "deactivate", "pause", "resume":
	default:
		return nil, errors.New("invalid action; must be deactivate/pause/resume")
	}

	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`INSERT INTO shop_lifecycle_request (shop_id, action, reason, status, admin_id, admin_remark, create_time, update_time)
		 VALUES (?, ?, ?, 0, 0, '', ?, ?)`,
		in.ShopId, in.Action, in.Reason, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &shop.SubmitShopLifecycleRequestResp{RequestId: id}, nil
}
