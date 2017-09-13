package execution

import (
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Execution The execution of a workflow
type Execution interface {
	Start()
	Complete()
	BuildStepImage(image string, options *image.BuildOptions) error
	RunStep(spec *RunStepSpec) error
	UpdateWorkflow(workflow *v1.Workflow, update func(*v1.Workflow)) error
}

// RunStepSpec Spec for a step to run
type RunStepSpec struct {
	Async            bool
	Command          []string
	Environment      *properties.Properties
	Image            string
	Name             string
	PodListener      kube.PodListener
	Readiness        *v1.HealthCheck
	VariableReceiver func(string, string)
	Volumes          []v1.Volume
}

// StepExecutionContext Context for a workflow step execution
type StepExecutionContext struct {
	Execution    Execution
	Workflow     *v1.Workflow
	StepSelector []int
	Step         *v1.WorkflowStep
}
