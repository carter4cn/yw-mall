package svc

import (
	"mall-activity-rpc/activityclient"
	"mall-order-rpc/orderclient"
	"mall-product-rpc/productclient"
	"mall-reward-rpc/rewardclient"
	"mall-rule-rpc/ruleclient"
	"mall-user-rpc/userclient"
	"mall-workflow-rpc/internal/config"
	"mall-workflow-rpc/internal/fsm"
	"mall-workflow-rpc/internal/model"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	DB                      sqlx.SqlConn
	Redis                   *redis.Redis
	AsynqClient             *asynq.Client
	Registry                *fsm.Registry
	Persister               *fsm.Persister
	WorkflowDefinitionModel model.WorkflowDefinitionModel
	WorkflowInstanceModel   model.WorkflowInstanceModel
	WorkflowStepLogModel    model.WorkflowStepLogModel
	AsynqTaskArchiveModel   model.AsynqTaskArchiveModel
	RuleRpc                 ruleclient.Rule
	RewardRpc               rewardclient.Reward
	ActivityRpc             activityclient.Activity
	UserRpc                 userclient.User
	ProductRpc              productclient.Product
	OrderRpc                orderclient.Order
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rds := redis.MustNewRedis(c.RedisCache)

	defModel := model.NewWorkflowDefinitionModel(conn, c.Cache)
	instModel := model.NewWorkflowInstanceModel(conn, c.Cache)
	stepModel := model.NewWorkflowStepLogModel(conn, c.Cache)

	asynqCli := asynq.NewClient(asynq.RedisClientOpt{
		Addr: c.RedisCache.Host,
		DB:   0,
	})

	return &ServiceContext{
		Config:                  c,
		DB:                      conn,
		Redis:                   rds,
		AsynqClient:             asynqCli,
		Registry:                fsm.NewRegistry(defModel),
		Persister:               fsm.NewPersister(conn, instModel, stepModel),
		WorkflowDefinitionModel: defModel,
		WorkflowInstanceModel:   instModel,
		WorkflowStepLogModel:    stepModel,
		AsynqTaskArchiveModel:   model.NewAsynqTaskArchiveModel(conn, c.Cache),
		RuleRpc:                 ruleclient.NewRule(zrpc.MustNewClient(c.RuleRpc)),
		RewardRpc:               rewardclient.NewReward(zrpc.MustNewClient(c.RewardRpc)),
		ActivityRpc:             activityclient.NewActivity(zrpc.MustNewClient(c.ActivityRpc)),
		UserRpc:                 userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc:              productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		OrderRpc:                orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
	}
}
