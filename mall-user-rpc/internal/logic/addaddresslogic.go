package logic

import (
	"context"
	"time"

	"mall-common/errorx"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AddAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddAddressLogic) AddAddress(in *user.AddAddressReq) (*user.AddAddressResp, error) {
	var newId int64
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		var cnt int64
		if e := sess.QueryRowCtx(ctx, &cnt, "SELECT COUNT(*) FROM user_address WHERE user_id=?", in.UserId); e != nil {
			return e
		}
		if cnt >= 20 {
			return errorx.NewCodeError(errorx.UserAddressLimit)
		}
		effectiveDefault := in.IsDefault || cnt == 0
		if effectiveDefault {
			if _, e := sess.ExecCtx(ctx, "UPDATE user_address SET is_default=0 WHERE user_id=?", in.UserId); e != nil {
				return e
			}
		}
		var isDefaultVal int64
		if effectiveDefault {
			isDefaultVal = 1
		}
		now := time.Now().Unix()
		res, e := sess.ExecCtx(ctx,
			"INSERT INTO user_address (user_id, receiver_name, phone, province, city, district, detail, is_default, create_time, update_time) VALUES (?,?,?,?,?,?,?,?,?,?)",
			in.UserId, in.ReceiverName, in.Phone, in.Province, in.City, in.District, in.Detail, isDefaultVal, now, now)
		if e != nil {
			return e
		}
		newId, e = res.LastInsertId()
		return e
	})
	if err != nil {
		return nil, err
	}
	return &user.AddAddressResp{Id: newId}, nil
}
