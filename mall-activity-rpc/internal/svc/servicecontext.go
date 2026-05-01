package svc

import (
	"mall-activity-rpc/internal/config"
	"mall-activity-rpc/internal/model"
	"mall-reward-rpc/rewardclient"
	"mall-risk-rpc/riskclient"
	"mall-rule-rpc/ruleclient"
	"mall-workflow-rpc/workflowclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                         config.Config
	DB                             sqlx.SqlConn
	Redis                          *redis.Redis
	ActivityModel                  model.ActivityModel
	ActivityTemplateModel          model.ActivityTemplateModel
	ParticipationRecordModel       model.ParticipationRecordModel
	ActivityStatModel              model.ActivityStatModel
	ActivityInventorySnapshotModel model.ActivityInventorySnapshotModel
	RuleRpc                        ruleclient.Rule
	WorkflowRpc                    workflowclient.Workflow
	RiskRpc                        riskclient.Risk
	RewardRpc                      rewardclient.Reward
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rds := redis.MustNewRedis(c.RedisCache)
	return &ServiceContext{
		Config:                         c,
		DB:                             conn,
		Redis:                          rds,
		ActivityModel:                  model.NewActivityModel(conn, c.Cache),
		ActivityTemplateModel:          model.NewActivityTemplateModel(conn, c.Cache),
		ParticipationRecordModel:       model.NewParticipationRecordModel(conn, c.Cache),
		ActivityStatModel:              model.NewActivityStatModel(conn, c.Cache),
		ActivityInventorySnapshotModel: model.NewActivityInventorySnapshotModel(conn, c.Cache),
		RuleRpc:                        ruleclient.NewRule(zrpc.MustNewClient(c.RuleRpc)),
		WorkflowRpc:                    workflowclient.NewWorkflow(zrpc.MustNewClient(c.WorkflowRpc)),
		RiskRpc:                        riskclient.NewRisk(zrpc.MustNewClient(c.RiskRpc)),
		RewardRpc:                      rewardclient.NewReward(zrpc.MustNewClient(c.RewardRpc)),
	}
}
