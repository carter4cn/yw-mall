package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAddressLogic) DeleteAddress(req *types.DeleteAddressReq) (*types.FollowShopResp, error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	_, err := l.svcCtx.UserRpc.DeleteAddress(l.ctx, &userclient.DeleteAddressReq{
		UserId: userId,
		Id:     req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.FollowShopResp{Ok: true}, nil
}
