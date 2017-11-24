package context

import (
	"context"
	"sync"

	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// StepContext Context for a step execution
type StepContext struct {
	WorkflowContext  *WorkflowContext
	Change           *v1.Change
	NextStep         *v1.WorkflowStep
	NextStepSelector []int
	Step             *v1.WorkflowStep
	StepSelector     []int
}

// WorkflowContext Context for a workflow execution
type WorkflowContext struct {
	Cancel   func()
	Cleanup  *sync.WaitGroup
	Context  context.Context
	Workflow *v1.Workflow
}
