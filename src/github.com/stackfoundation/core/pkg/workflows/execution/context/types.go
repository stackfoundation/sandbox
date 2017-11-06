package context

import (
	"context"

	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// ExecutionContext Context for an execution
type ExecutionContext struct {
	Context          context.Context
	Change           *v1.Change
	NextStep         *v1.WorkflowStep
	NextStepSelector []int
	Step             *v1.WorkflowStep
	StepSelector     []int
	Workflow         *v1.Workflow
}
