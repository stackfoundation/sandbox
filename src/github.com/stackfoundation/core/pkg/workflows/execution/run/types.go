package run

import (
	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Result Stores the result of a pod step run
type Result struct {
	Container string
	Message   string
	Workflow  string
	Variables []v1.VariableSource
}

// Listener Listens to a pod step run
type Listener interface {
	Ready(sc *context.StepContext)
	Failed(sc *context.StepContext, r *Result)
	Done(sc *context.StepContext, r *Result)
}

type podCompletionListener struct {
	listener           Listener
	stepContext        *context.StepContext
	generatedContainer string
	generatedWorkflow  string
	variables          []v1.VariableSource
}
