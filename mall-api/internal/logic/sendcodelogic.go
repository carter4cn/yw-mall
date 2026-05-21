// mall-api/internal/logic/sendcodelogic.go
package logic

import (
	"context"
	"errors"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

func SendCode(ctx context.Context, svcCtx *svc.ServiceContext, req *types.SendCodeReq) (*types.SendCodeResp, error) {
	if req.Target == "" {
		return nil, errors.New("target required")
	}
	resp, err := svcCtx.UserRpc.SendVerifyCode(ctx, &userclient.SendVerifyCodeReq{
		Channel: req.Channel, Target: req.Target, Scene: req.Scene,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("SendCode rpc fail: %v", err)
		return nil, err
	}
	return &types.SendCodeResp{
		ChallengeToken: resp.ChallengeToken,
		ExpiresIn:      resp.ExpiresIn,
	}, nil
}
