package execution

import (
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Execution The execution of a workflow
type Execution interface {
	ChildExecution(workflow *v1.Workflow) (Execution, error)
	Start()
	Complete() error
	BuildStepImage(image string, options *image.BuildOptions) error
	RunStep(spec *RunStepSpec) error
	TransitionNext(context *Context, update func(*Context, *v1.Workflow)) error
}

// RunStepSpec Spec for a step to run
type RunStepSpec struct {
	Command          []string
	Environment      *properties.Properties
	Image            string
	Name             string
	PodListener      kube.PodListener
	Readiness        *v1.HealthCheck
	VariableReceiver func(string, string)
	Volumes          []v1.Volume
	WorkflowReceiver func(string)
}

// Context Context for an execution
type Context struct {
	Workflow         *v1.Workflow
	Change           *v1.Change
	Step             *v1.WorkflowStep
	StepSelector     []int
	NextStep         *v1.WorkflowStep
	NextStepSelector []int
}

// NewContext Create a new context with the given workflow and change
func NewContext(w *v1.Workflow, c *v1.Change) *Context {
	nextStepSelector := w.IncrementStepSelector(c.StepSelector)
	return &Context{
		Workflow:         w,
		Change:           c,
		Step:             w.Select(c.StepSelector),
		StepSelector:     c.StepSelector,
		NextStep:         w.Select(nextStepSelector),
		NextStepSelector: nextStepSelector,
	}
}

// IsCompoundStepBoundary Does the next step in this context exit a compound step?
func (c *Context) IsCompoundStepBoundary() bool {
	currentSegmentCount := len(c.StepSelector)
	nextSegmentCount := len(c.NextStepSelector)

	return nextSegmentCount < currentSegmentCount
}

// IsWorkflowComplete Is the context at the end of the workflow?
func (c *Context) IsWorkflowComplete() bool {
	nextSegmentCount := len(c.NextStepSelector)
	return nextSegmentCount == 0
}
