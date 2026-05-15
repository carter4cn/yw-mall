package svc

import (
	"mall-common/configcenter"
	"mall-common/cryptox"
	"mall-user-rpc/internal/config"
	"mall-user-rpc/internal/model"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gopkg.in/yaml.v3"
)

// PasswordPolicy is the in-memory policy used by validatePassword,
// recordPasswordHistory and the login expiry check (S4.3). Defaults are set in
// NewServiceContext; future Sprint 5 work moves these to etcd hot-reload.
type PasswordPolicy struct {
	MinLength     int
	RequireUpper  bool
	RequireLower  bool
	RequireDigit  bool
	RequireSymbol bool
	MaxHistory    int
	MaxAgeDays    int
}

type ServiceContext struct {
	Config           config.Config
	DB               sqlx.SqlConn
	UserModel        model.UserModel
	UserAddressModel model.UserAddressModel

	// JwtSecretHot is hot-reloadable: kept for backward compatibility with the
	// old JWT path (admin login still inspects it). P0 login revamp swaps the
	// user-facing path to opaque tokens, so this field becomes a no-op for
	// user-side flows.
	JwtSecretHot *configcenter.HotConfig[string]

	// Redis holds opaque-token sessions: session:{access}, refresh:{refresh},
	// and user_sessions:{uid} index. See logic/sessionhelpers.go for layout.
	Redis *redis.Client

	// PasswordPolicy: S4.3 strength + history + expiry policy. Built from
	// defaults; configurable via etcd in Sprint 5.
	PasswordPolicy PasswordPolicy
}

func NewServiceContext(c config.Config, etcdHosts []string) *ServiceContext {
	// S4.6: fail-fast if MALL_FIELD_ENCRYPTION_KEY is missing/invalid. We rely
	// on cryptox for new user.phone writes and admin_mfa.totp_secret_enc, so a
	// silent miss would corrupt every new row.
	cryptox.MustInit()

	conn := sqlx.NewMysql(c.DataSource)
	svc := &ServiceContext{
		Config:           c,
		DB:               conn,
		UserModel:        model.NewUserModel(conn, c.Cache),
		UserAddressModel: model.NewUserAddressModel(conn, c.Cache),
		JwtSecretHot:     configcenter.NewHotConfig(c.JwtAuth.AccessSecret),
		Redis:            newRedisClient(c),
		PasswordPolicy: PasswordPolicy{
			MinLength:     8,
			RequireUpper:  true,
			RequireLower:  true,
			RequireDigit:  true,
			RequireSymbol: false,
			MaxHistory:    5,
			MaxAgeDays:    90,
		},
	}
	if len(etcdHosts) > 0 {
		go configcenter.NewWatcher(etcdHosts).Watch(configcenter.ServiceKey("yw-mall", "user-rpc"), svc.onConfigChange)
	}
	return svc
}

func newRedisClient(c config.Config) *redis.Client {
	host := c.Session.Redis.Host
	if host == "" {
		// Fallback to the first Cache node so an unconfigured deployment still works.
		if len(c.Cache) > 0 {
			host = c.Cache[0].Host
		}
	}
	return redis.NewClient(&redis.Options{
		Addr:     host,
		Password: c.Session.Redis.Pass,
		DB:       c.Session.Redis.DB,
	})
}

func (s *ServiceContext) onConfigChange(data []byte) {
	var newCfg config.Config
	if err := yaml.Unmarshal(data, &newCfg); err != nil {
		logx.Errorf("[configcenter] user-rpc config parse error: %v", err)
		return
	}
	s.JwtSecretHot.Set(newCfg.JwtAuth.AccessSecret)
	logx.Infof("[configcenter] user-rpc: JwtAuth.AccessSecret updated")
}
