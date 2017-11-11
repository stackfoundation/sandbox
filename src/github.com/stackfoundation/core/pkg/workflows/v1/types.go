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

// HealthCheckOptions Base options for health checks
type HealthCheckOptions struct {
	Grace    string `json:"grace" yaml:"grace"`
	Interval string `json:"interval" yaml:"interval"`
	Retries  string `json:"retries" yaml:"retries"`
	SkipWait string `json:"skipWait" yaml:"skipWait"`
	Timeout  string `json:"timeout" yaml:"timeout"`
}

// TCPHealthCheckOptions Options for TCP health checks
type TCPHealthCheckOptions struct {
	HealthCheckOptions `json:",inline" yaml:",inline"`

	Port string `json:"port" yaml:"port"`
}

// ScriptHealthCheckOptions Options for script health checks
type ScriptHealthCheckOptions struct {
	HealthCheckOptions `json:",inline" yaml:",inline"`

	Path string `json:"path" yaml:"path"`
}

// HTTPHealthCheckOptions Options for HTTP/HTTPS health checks
type HTTPHealthCheckOptions struct {
	HealthCheckOptions `json:",inline" yaml:",inline"`

	Headers []HTTPHeader `json:"headers" yaml:"headers"`
	Path    string       `json:"path" yaml:"path"`
	Port    string       `json:"port" yaml:"port"`
}

// HealthCheck Health check for a workflow step
type HealthCheck struct {
	HTTP   *HTTPHealthCheckOptions   `json:"http" yaml:"http"`
	HTTPS  *HTTPHealthCheckOptions   `json:"https" yaml:"https"`
	TCP    *TCPHealthCheckOptions    `json:"tcp" yaml:"tcp"`
	Script *ScriptHealthCheckOptions `json:"script" yaml:"script"`
}

// StepState State of step
type StepState struct {
	GeneratedBaseImage string `json:"baseImage" yaml:"baseImage"`
	GeneratedImage     string `json:"generatedImage" yaml:"generatedImage"`
	GeneratedContainer string `json:"generatedContainer" yaml:"generatedContainer"`
	GeneratedScript    string `json:"generatedScript" yaml:"generatedScript"`
	GeneratedWorkflow  string `json:"generatedWorkflow" yaml:"generatedWorkflow"`
	Ready              bool   `json:"ready" yaml:"ready"`
	Done               bool   `json:"done" yaml:"done"`
	Prepared           bool   `json:"prepared" yaml:"prepared"`
}

// Port An exposed port
type Port struct {
	Protocol  string `json:"protocol" yaml:"protocol"`
	Name      string `json:"name" yaml:"name"`
	Container string `json:"container" yaml:"container"`
	External  string `json:"external" yaml:"external"`
	Internal  string `json:"internal" yaml:"internal"`
}

// SourceOptions Source options for a step
type SourceOptions struct {
	Dockerignore string   `json:"dockerignore" yaml:"dockerignore"`
	Exclude      []string `json:"exclude" yaml:"exclude"`
	Include      []string `json:"include" yaml:"include"`
	Location     string   `json:"location" yaml:"location"`
	Omit         string   `json:"omit" yaml:"omit"`
}

// VariableOptions Variable options for a step
type VariableOptions struct {
	Exclude []string `json:"exclude" yaml:"exclude"`
	Include []string `json:"include" yaml:"include"`
}

// StepOptions Options for all types of steps
type StepOptions struct {
	Name             string `json:"name" yaml:"name"`
	IgnoreFailure    *bool  `json:"ignoreFailure" yaml:"ignoreFailure"`
	IgnoreMissing    *bool  `json:"ignoreMissing" yaml:"ignoreMissing"`
	IgnoreValidation *bool  `json:"ignoreValidation" yaml:"ignoreValidation"`
}

// ExternalStepOptions Options for external steps
type ExternalStepOptions struct {
	StepOptions `json:",inline" yaml:",inline"`

	Parallel  string          `json:"parallel" yaml:"parallel"`
	Variables VariableOptions `json:"variables" yaml:"variables"`
	Workflow  string          `json:"workflow" yaml:"workflow"`
}

// CompoundStepOptions Options for compound steps
type CompoundStepOptions struct {
	StepOptions `json:",inline" yaml:",inline"`

	Steps []WorkflowStep `json:"steps" yaml:"steps"`
}

// ScriptStepOptions Options for script-based steps
type ScriptStepOptions struct {
	StepOptions `json:",inline" yaml:",inline"`

	Dockerfile  string           `json:"dockerfile" yaml:"dockerfile"`
	Environment []VariableSource `json:"environment" yaml:"environment"`
	Image       string           `json:"image" yaml:"image"`
	Script      string           `json:"script" yaml:"script"`
	Source      SourceOptions    `json:"source" yaml:"source"`
	Step        string           `json:"step" yaml:"step"`
	Volumes     []Volume         `json:"volumes" yaml:"volumes"`
}

// ServiceStepOptions Options for a service step
type ServiceStepOptions struct {
	ScriptStepOptions `json:",inline" yaml:",inline"`

	Grace     string       `json:"grace" yaml:"grace"`
	Health    *HealthCheck `json:"health" yaml:"health"`
	Ports     []Port       `json:"ports" yaml:"ports"`
	Readiness *HealthCheck `json:"readiness" yaml:"readiness"`
}

// GeneratorStepOptions Options for a generator step
type GeneratorStepOptions struct {
	ScriptStepOptions `json:",inline" yaml:",inline"`

	Parallel  string          `json:"parallel" yaml:"parallel"`
	Variables VariableOptions `json:"variables" yaml:"variables"`
}

// RunStepOptions Options for a run step
type RunStepOptions struct {
	ScriptStepOptions `json:",inline" yaml:",inline"`

	Cache    string `json:"cache" yaml:"cache"`
	Parallel string `json:"parallel" yaml:"parallel"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	Compound  *CompoundStepOptions  `json:"compound" yaml:"compound"`
	External  *ExternalStepOptions  `json:"external" yaml:"external"`
	Generator *GeneratorStepOptions `json:"generator" yaml:"generator"`
	Run       *RunStepOptions       `json:"run" yaml:"run"`
	Service   *ServiceStepOptions   `json:"service" yaml:"service"`
	State     StepState             `json:"state" yaml:"state"`
}

// StepStarted A step was started
const StepStarted ChangeType = "stepStarted"

// StepReady A parallel or service step is ready
const StepReady ChangeType = "stepReady"

// StepDone A parallel or service step is done
const StepDone ChangeType = "stepDone"

// WorkflowWait A step is waiting for a workflow
const WorkflowWait ChangeType = "workflowWait"

// WorkflowWaitDone Wait for a workflow to complete is done
const WorkflowWaitDone ChangeType = "workflowWaitDone"

// StepImageBuilt Image for step has been built
const StepImageBuilt ChangeType = "stepImageBuilt"

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
	Variables   *properties.Properties `json:"-" yaml:"-"`
	Changes     []Change               `json:"changes" yaml:"changes"`
	Step        []int                  `json:"step" yaml:"step"`
}

// WorkflowSpec Specification of workflow
type WorkflowSpec struct {
	State            WorkflowState    `json:"state" yaml:"state"`
	Steps            []WorkflowStep   `json:"steps" yaml:"steps"`
	Variables        []VariableSource `json:"variables" yaml:"variables"`
	IgnoreMissing    bool             `json:"ignoreMissing" yaml:"ignoreMissing"`
	IgnoreValidation bool             `json:"ignoreValidation" yaml:"ignoreValidation"`
	IgnoreFailure    bool             `json:"ignoreFailure" yaml:"ignoreFailure"`
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

// SchemeGroupVersion Workflows GroupVersion
var SchemeGroupVersion = schema.GroupVersion{
	Group:   WorkflowsGroupName,
	Version: WorkflowsGroupVersion,
}
