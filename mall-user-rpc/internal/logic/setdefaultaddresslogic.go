package logic

import (
	"context"

	"mall-common/errorx"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SetDefaultAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDefaultAddressLogic) SetDefaultAddress(in *user.SetDefaultAddressReq) (*user.OkResp, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		if _, e := sess.ExecCtx(ctx, "UPDATE user_address SET is_default=0 WHERE user_id=?", in.UserId); e != nil {
			return e
		}
		res, e := sess.ExecCtx(ctx, "UPDATE user_address SET is_default=1 WHERE id=? AND user_id=?", in.Id, in.UserId)
		if e != nil {
			return e
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return errorx.NewCodeError(errorx.UserAddressForbidden)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}
