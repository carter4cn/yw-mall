package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WorkflowStepLogModel = (*customWorkflowStepLogModel)(nil)

type (
	// WorkflowStepLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWorkflowStepLogModel.
	WorkflowStepLogModel interface {
		workflowStepLogModel
	}

	customWorkflowStepLogModel struct {
		*defaultWorkflowStepLogModel
	}
)

// NewWorkflowStepLogModel returns a model for the database table.
func NewWorkflowStepLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WorkflowStepLogModel {
	return &customWorkflowStepLogModel{
		defaultWorkflowStepLogModel: newWorkflowStepLogModel(conn, c, opts...),
	}
}
