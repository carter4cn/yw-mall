package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WorkflowDefinitionModel = (*customWorkflowDefinitionModel)(nil)

type (
	// WorkflowDefinitionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWorkflowDefinitionModel.
	WorkflowDefinitionModel interface {
		workflowDefinitionModel
	}

	customWorkflowDefinitionModel struct {
		*defaultWorkflowDefinitionModel
	}
)

// NewWorkflowDefinitionModel returns a model for the database table.
func NewWorkflowDefinitionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WorkflowDefinitionModel {
	return &customWorkflowDefinitionModel{
		defaultWorkflowDefinitionModel: newWorkflowDefinitionModel(conn, c, opts...),
	}
}
