package logic

import (
	"context"

	"mall-common/errorx"
	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAddressLogic) DeleteAddress(in *user.DeleteAddressReq) (*user.OkResp, error) {
	addr, err := l.svcCtx.UserAddressModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.UserAddressNotFound)
		}
		return nil, err
	}
	if int64(addr.UserId) != in.UserId {
		return nil, errorx.NewCodeError(errorx.UserAddressForbidden)
	}
	wasDefault := addr.IsDefault == 1
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		if _, e := sess.ExecCtx(ctx, "DELETE FROM user_address WHERE id=?", in.Id); e != nil {
			return e
		}
		if wasDefault {
			var nextId uint64
			e := sess.QueryRowCtx(ctx, &nextId,
				"SELECT id FROM user_address WHERE user_id=? ORDER BY update_time DESC LIMIT 1", in.UserId)
			if e == sqlx.ErrNotFound {
				return nil
			}
			if e != nil {
				return e
			}
			_, e = sess.ExecCtx(ctx, "UPDATE user_address SET is_default=1 WHERE id=?", nextId)
			return e
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = l.svcCtx.UserAddressModel.Delete(l.ctx, uint64(in.Id))
	return &user.OkResp{Ok: true}, nil
}
