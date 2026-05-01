package svc

import (
	rcache "mall-rule-rpc/internal/cache"
	"mall-rule-rpc/internal/config"
	"mall-rule-rpc/internal/loader"
	"mall-rule-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                 config.Config
	RuleModel              model.RuleModel
	RuleSetModel           model.RuleSetModel
	RuleEvaluationLogModel model.RuleEvaluationLogModel
	ProgramCache           *rcache.ProgramCache
	Loader                 *loader.Loader
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rm := model.NewRuleModel(conn, c.Cache)
	pc, err := rcache.NewProgramCache(c.LruCacheSize)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                 c,
		RuleModel:              rm,
		RuleSetModel:           model.NewRuleSetModel(conn, c.Cache),
		RuleEvaluationLogModel: model.NewRuleEvaluationLogModel(conn, c.Cache),
		ProgramCache:           pc,
		Loader:                 loader.New(rm, pc),
	}
}
