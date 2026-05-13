package logic

import (
	"context"

	"mall-api/internal/middleware"
	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAddressLogic) UpdateAddress(req *types.UpdateAddressReq) (*types.FollowShopResp, error) {
	userId := middleware.UidFromCtx(l.ctx)

	_, err := l.svcCtx.UserRpc.UpdateAddress(l.ctx, &userclient.UpdateAddressReq{
		UserId:       userId,
		Id:           req.Id,
		ReceiverName: req.ReceiverName,
		Phone:        req.Phone,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Detail:       req.Detail,
	})
	if err != nil {
		return nil, err
	}
	return &types.FollowShopResp{Ok: true}, nil
}
