package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/userclient"
)

func ResetPassword(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ResetPasswordReq) error {
	_, err := svcCtx.UserRpc.ResetPasswordByCode(ctx, &userclient.ResetPasswordByCodeReq{
		Phone:          req.Phone,
		VerifyCode:     req.VerifyCode,
		ChallengeToken: req.ChallengeToken,
		NewPassword:    req.NewPassword,
	})
	return err
}
