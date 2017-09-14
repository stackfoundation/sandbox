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
	Change           *v1.Change
	NextStep         *v1.WorkflowStep
	NextStepSelector []int
	Step             *v1.WorkflowStep
	StepSelector     []int
	Workflow         *v1.Workflow
}
