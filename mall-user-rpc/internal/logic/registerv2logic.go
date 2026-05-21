// mall-user-rpc/internal/logic/registerv2logic.go
package logic

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"mall-common/cryptox"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterV2Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterV2Logic {
	return &RegisterV2Logic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

const (
	sceneRegister = 1
)

var (
	usernameRE = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{3,31}$`) // 4-32, 字母开头
	phoneRE    = regexp.MustCompile(`^\d{11}$`)
	emailRE    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func (l *RegisterV2Logic) RegisterV2(in *user.RegisterV2Req) (*user.RegisterResp, error) {
	username := strings.TrimSpace(in.Username)
	phone := strings.TrimSpace(in.Phone)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	password := trimPlain(in.Password)

	if !usernameRE.MatchString(username) {
		return nil, errors.New("用户名格式不合法（4-32 位字母开头）")
	}
	if phone == "" && email == "" {
		return nil, errors.New("手机号和邮箱至少填一个")
	}
	if phone != "" && !phoneRE.MatchString(phone) {
		return nil, errors.New("手机号格式不合法")
	}
	if email != "" && !emailRE.MatchString(email) {
		return nil, errors.New("邮箱格式不合法")
	}
	if err := validatePassword(password, l.svcCtx.PasswordPolicy); err != nil {
		return nil, err
	}

	// 1. username 唯一
	var cnt int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &cnt,
		"SELECT COUNT(*) FROM `user` WHERE username=?", username); err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, errors.New("用户名已被使用")
	}

	// 2. 验证码消费 + identifier 占用检查
	target := phone
	if target == "" {
		target = email
	}
	if err := consumeVerifyCode(l.ctx, l.svcCtx, sceneRegister, target,
		in.VerifyCode, in.ChallengeToken); err != nil {
		return nil, err
	}

	phoneHash := cryptox.Hmac(phone)
	emailHash := cryptox.Hmac(email)

	if phoneHash != "" {
		if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &cnt,
			"SELECT COUNT(*) FROM `user` WHERE phone_hash=?", phoneHash); err != nil {
			return nil, err
		}
		if cnt > 0 {
			return nil, errors.New("手机号已被注册")
		}
	}
	if emailHash != "" {
		if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &cnt,
			"SELECT COUNT(*) FROM `user` WHERE email_hash=?", emailHash); err != nil {
			return nil, err
		}
		if cnt > 0 {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 3. 加密 + bcrypt
	var phoneEnc, emailEnc string
	if phone != "" {
		e, err := cryptox.Encrypt(phone)
		if err != nil {
			return nil, err
		}
		phoneEnc = e
	}
	if email != "" {
		e, err := cryptox.Encrypt(email)
		if err != nil {
			return nil, err
		}
		emailEnc = e
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. INSERT
	now := time.Now().Unix()
	res, err := l.svcCtx.DB.ExecCtx(l.ctx, `
        INSERT INTO user (username, password, phone, email, phone_enc, email_enc,
                          phone_hash, email_hash, last_password_change)
        VALUES (?, ?, '', ?, ?, ?, ?, ?, ?)`,
		username, string(hash), email, phoneEnc, emailEnc, phoneHash, emailHash, now)
	if err != nil {
		if isDuplicateKeyErr(err) {
			return nil, errors.New("用户名/手机号/邮箱已被使用")
		}
		return nil, err
	}
	id, _ := res.LastInsertId()

	// 5. 记一次 password_history
	_ = recordPasswordHistory(l.ctx, l.svcCtx.DB, subjectTypeUser, uint64(id),
		string(hash), l.svcCtx.PasswordPolicy.MaxHistory)

	return &user.RegisterResp{Id: id}, nil
}

func isDuplicateKeyErr(err error) bool {
	if err == sqlx.ErrNotFound {
		return false
	}
	return strings.Contains(err.Error(), "Duplicate entry")
}
