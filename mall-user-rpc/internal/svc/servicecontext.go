package svc

import (
	"mall-common/configcenter"
	"mall-user-rpc/internal/config"
	"mall-user-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gopkg.in/yaml.v3"
)

type ServiceContext struct {
	Config           config.Config
	DB               sqlx.SqlConn
	UserModel        model.UserModel
	UserAddressModel model.UserAddressModel

	// JwtSecretHot is hot-reloadable: updated by etcd watcher, read by LoginLogic.
	JwtSecretHot *configcenter.HotConfig[string]
}

func NewServiceContext(c config.Config, etcdHosts []string) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	svc := &ServiceContext{
		Config:           c,
		DB:               conn,
		UserModel:        model.NewUserModel(conn, c.Cache),
		UserAddressModel: model.NewUserAddressModel(conn, c.Cache),
		JwtSecretHot:     configcenter.NewHotConfig(c.JwtAuth.AccessSecret),
	}
	if len(etcdHosts) > 0 {
		go configcenter.NewWatcher(etcdHosts).Watch(configcenter.ServiceKey("yw-mall", "user-rpc"), svc.onConfigChange)
	}
	return svc
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
