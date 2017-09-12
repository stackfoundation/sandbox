package execution

import (
	"github.com/magiconair/properties"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
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
	Name        string
	Image       string
	Command     []string
	Volumes     []v1.Volume
	Readiness   *v1.HealthCheck
	Environment *properties.Properties
	Updater     kube.PodStatusUpdater
}

// StepExecutionContext Context for a workflow step execution
type StepExecutionContext struct {
	Execution    Execution
	Workflow     *v1.Workflow
	StepSelector []int
	Step         *v1.WorkflowStep
}
