// mall-api/internal/logic/registerv2logic.go
package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/userclient"
)

func RegisterV2(ctx context.Context, svcCtx *svc.ServiceContext, req *types.RegisterV2Req) (*types.RegisterV2Resp, error) {
	resp, err := svcCtx.UserRpc.RegisterV2(ctx, &userclient.RegisterV2Req{
		Username:       req.Username,
		Password:       req.Password,
		Phone:          req.Phone,
		Email:          req.Email,
		VerifyCode:     req.VerifyCode,
		ChallengeToken: req.ChallengeToken,
	})
	if err != nil {
		return nil, err
	}
	return &types.RegisterV2Resp{Id: resp.Id}, nil
}
