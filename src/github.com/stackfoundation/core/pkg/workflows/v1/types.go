package v1

import (
	"github.com/stackfoundation/core/pkg/workflows/properties"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// VariableSource A source of a variable
type VariableSource struct {
	File  string `json:"file" yaml:"file"`
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// Volume Volume to mount for a workflow step
type Volume struct {
	HostPath  string `json:"hostPath" yaml:"hostPath"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	Name      string `json:"name" yaml:"name"`
}

// HTTPHeader HTTP header to send in health check
type HTTPHeader struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// HealthCheckType Health check type
type HealthCheckType string

// TCPCheck TCP health check
const TCPCheck HealthCheckType = "tcp"

// HTTPCheck HTTP health check
const HTTPCheck HealthCheckType = "http"

// HTTPSCheck HTTPS health check
const HTTPSCheck HealthCheckType = "https"

// ScriptCheck Script health check
const ScriptCheck HealthCheckType = "script"

// HealthCheck HealthCheck checks for a workflow step
type HealthCheck struct {
	Grace    *int32          `json:"grace" yaml:"grace"`
	Headers  []HTTPHeader    `json:"headers" yaml:"headers"`
	Interval *int32          `json:"interval" yaml:"interval"`
	Path     string          `json:"path" yaml:"path"`
	Port     int32           `json:"port" yaml:"port"`
	Retries  *int32          `json:"retries" yaml:"retries"`
	Script   string          `json:"script" yaml:"script"`
	SkipWait bool            `json:"skipWait" yaml:"skipWait"`
	Timeout  *int32          `json:"timeout" yaml:"timeout"`
	Type     HealthCheckType `json:"type" yaml:"type"`
}

// ImageSource Image source
type ImageSource string

// SourceImage Docker image
const SourceImage ImageSource = "image"

// SourceStep Source is previous step image
const SourceStep ImageSource = "step"

// StepState State of step
type StepState struct {
	GeneratedImage    string `json:"generatedImage" yaml:"generatedImage"`
	GeneratedScript   string `json:"generatedScript" yaml:"generatedScript"`
	GeneratedWorkflow string `json:"generatedWorkflow" yaml:"generatedWorkflow"`
	Ready             bool   `json:"ready" yaml:"ready"`
	Done              bool   `json:"done" yaml:"done"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	Dockerfile       string           `json:"dockerfile" yaml:"dockerfile"`
	Environment      []VariableSource `json:"environment" yaml:"environment"`
	Generator        string           `json:"generator" yaml:"generator"`
	Health           *HealthCheck     `json:"health" yaml:"health"`
	Image            string           `json:"image" yaml:"image"`
	ImageSource      ImageSource      `json:"imageSource" yaml:"imageSource"`
	Name             string           `json:"name" yaml:"name"`
	OmitSource       bool             `json:"omitSource" yaml:"omitSource"`
	Ports            []string         `json:"ports" yaml:"ports"`
	Readiness        *HealthCheck     `json:"readiness" yaml:"readiness"`
	Script           string           `json:"script" yaml:"script"`
	SourceLocation   string           `json:"sourceLocation" yaml:"sourceLocation"`
	State            StepState        `json:"state" yaml:"state"`
	Steps            []WorkflowStep   `json:"steps" yaml:"steps"`
	Target           string           `json:"target" yaml:"target"`
	TerminationGrace *int32           `json:"terminationGrace" yaml:"terminationGrace"`
	Type             string           `json:"type" yaml:"type"`
	Volumes          []Volume         `json:"volumes" yaml:"volumes"`
}

// IsAsync Is this an async step (a paralell or service step that skips wait)?
func (s *WorkflowStep) IsAsync() bool {
	return s.Type == StepParallel ||
		(s.Type == StepService && s.Readiness != nil && s.Readiness.SkipWait) ||
		(len(s.Type) == 0 && s.Readiness != nil && s.Readiness.SkipWait)
}

// IsServiceWithWait Is this a service step that waits for readiness?
func (s *WorkflowStep) IsServiceWithWait() bool {
	return (s.Type == StepService && s.Readiness != nil && !s.Readiness.SkipWait) ||
		((len(s.Type) == 0) && s.Readiness != nil && !s.Readiness.SkipWait)
}

// RequiresBuild Does the step require an image to be built (all except call steps)?
func (s *WorkflowStep) RequiresBuild() bool {
	return len(s.Script) > 0 || len(s.Generator) > 0 || len(s.Dockerfile) > 0
}

// AsyncStepStarted An async step was started
const AsyncStepStarted ChangeType = "asyncStarted"

// StepStarted A step was started
const StepStarted ChangeType = "started"

// StepReady A parallel or service step is ready
const StepReady ChangeType = "ready"

// StepDone A parallel or service step is done
const StepDone ChangeType = "done"

// WorkflowWait A step is waiting for a workflow
const WorkflowWait ChangeType = "workflowWait"

// StepImageBuilt Image for step has been built
const StepImageBuilt ChangeType = "imageBuilt"

// ChangeType Type of change
type ChangeType string

// Change A workflow change
type Change struct {
	ID           string     `json:"id" yaml:"id"`
	Type         ChangeType `json:"type" yaml:"type"`
	Handled      bool       `json:"handled" yaml:"handled"`
	StepSelector []int      `json:"step" yaml:"step"`
}

// NewChange Create a new unhandled change, with a generated ID
func NewChange(selector []int) *Change {
	return &Change{
		ID:           GenerateChangeID(),
		StepSelector: selector,
	}
}

// WorkflowState State of workflow in K8s
type WorkflowState struct {
	ID          string                 `json:"id" yaml:"id"`
	ProjectRoot string                 `json:"projectRoot" yaml:"projectRoot"`
	Properties  *properties.Properties `json:"-" yaml:"-"`
	Changes     []Change               `json:"changes" yaml:"changes"`
	Step        []int                  `json:"step" yaml:"step"`
}

// WorkflowSpec Specification of workflow
type WorkflowSpec struct {
	State     WorkflowState    `json:"state" yaml:"state"`
	Steps     []WorkflowStep   `json:"steps" yaml:"steps"`
	Variables []VariableSource `json:"variables" yaml:"variables"`
}

// Workflow Custom workflow resource
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WorkflowSpec `json:"spec"`
}

// WorkflowList List of workflows
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Workflow `json:"items"`
}

// WorkflowsGroupName Group name for workflows
const WorkflowsGroupName = "stack.foundation"

// WorkflowsGroupVersion Group version for workflows
const WorkflowsGroupVersion = "v1"

// WorkflowsKind Kind for workflows
const WorkflowsKind = "Workflow"

// WorkflowsPluralName Plural form of name for workflows
const WorkflowsPluralName = "workflows"

// WorkflowsSingularName Singular form of name for workflows
const WorkflowsSingularName = "workflows"

// WorkflowsCustomResource Name of custom resource for workflows
const WorkflowsCustomResource = WorkflowsPluralName + "." + WorkflowsGroupName

// DefaultGrace Default grace period in seconds
const DefaultGrace = 0

// DefaultInterval Default interval in seconds
const DefaultInterval = 10

// DefaultRetries Default number of retries
const DefaultRetries = 3

// DefaultTimeout Default timeout in seconds
const DefaultTimeout = 10

// StepSequential Sequential step
const StepSequential = "sequential"

// StepParallel Parallel step
const StepParallel = "parallel"

// StepCompound Compound step
const StepCompound = "compound"

// StepService Service step
const StepService = "service"

// SchemeGroupVersion Workflows GroupVersion
var SchemeGroupVersion = schema.GroupVersion{
	Group:   WorkflowsGroupName,
	Version: WorkflowsGroupVersion,
}
