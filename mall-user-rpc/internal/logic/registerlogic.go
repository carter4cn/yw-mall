package logic

import (
	"context"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// S4.3 password policy: validate strength on registration. No history check
	// on first password.
	if err := validatePassword(in.Password, l.svcCtx.PasswordPolicy); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// S4.6 encrypt PII at rest. Empty phone is preserved as empty.
	phoneEnc, err := cryptox.Encrypt(in.Phone)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username: in.Username,
		Password: string(hash),
		Phone:    phoneEnc,
	})
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	// S4.3 also bump last_password_change so MaxAgeDays expiry tracking begins
	// at registration. Best-effort: a failure here doesn't roll back the user.
	now := time.Now().Unix()
	_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `user` SET last_password_change=? WHERE id=?", now, id)
	// Record password history for the new user (subject_type=1=user).
	_ = recordPasswordHistory(l.ctx, l.svcCtx.DB, subjectTypeUser, uint64(id), string(hash), l.svcCtx.PasswordPolicy.MaxHistory)

	return &user.RegisterResp{Id: id}, nil
}
