package controller

import (
	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Controller A controller used to execute workflows
type Controller interface {
	Execute(workflow *v1.Workflow) error
}

type executionController struct {
	coordinator        coordinator.Coordinator
	pendingTransitions chan pendingTransition
}

type pendingTransition struct {
	context    *context.Context
	transition func(*context.Context, *v1.Workflow)
}
