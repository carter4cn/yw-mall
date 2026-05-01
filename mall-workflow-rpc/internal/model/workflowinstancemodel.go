package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WorkflowInstanceModel = (*customWorkflowInstanceModel)(nil)

type (
	// WorkflowInstanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWorkflowInstanceModel.
	WorkflowInstanceModel interface {
		workflowInstanceModel
	}

	customWorkflowInstanceModel struct {
		*defaultWorkflowInstanceModel
	}
)

// NewWorkflowInstanceModel returns a model for the database table.
func NewWorkflowInstanceModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WorkflowInstanceModel {
	return &customWorkflowInstanceModel{
		defaultWorkflowInstanceModel: newWorkflowInstanceModel(conn, c, opts...),
	}
}
