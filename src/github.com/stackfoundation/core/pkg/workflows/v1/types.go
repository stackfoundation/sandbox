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
	Grace    string          `json:"grace" yaml:"grace"`
	Headers  []HTTPHeader    `json:"headers" yaml:"headers"`
	Interval string          `json:"interval" yaml:"interval"`
	Path     string          `json:"path" yaml:"path"`
	Port     string          `json:"port" yaml:"port"`
	Retries  string          `json:"retries" yaml:"retries"`
	Script   string          `json:"script" yaml:"script"`
	SkipWait string          `json:"skipWait" yaml:"skipWait"`
	Timeout  string          `json:"timeout" yaml:"timeout"`
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
	Protocol      string `json:"protocol" yaml:"protocol"`
	Name          string `json:"name" yaml:"name"`
	ContainerPort string `json:"containerPort" yaml:"containerPort"`
	ExternalPort  string `json:"externalPort" yaml:"externalPort"`
	InternalPort  string `json:"internalPort" yaml:"internalPort"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	Cache            string           `json:"cache" yaml:"cache"`
	Dockerfile       string           `json:"dockerfile" yaml:"dockerfile"`
	Dockerignore     string           `json:"dockerignore" yaml:"dockerignore"`
	Environment      []VariableSource `json:"environment" yaml:"environment"`
	ExcludeVariables []string         `json:"excludeVariables" yaml:"excludeVariables"`
	Generator        string           `json:"generator" yaml:"generator"`
	Health           *HealthCheck     `json:"health" yaml:"health"`
	Image            string           `json:"image" yaml:"image"`
	IgnoreFailure    *bool            `json:"ignoreFailure" yaml:"ignoreFailure"`
	IgnoreMissing    *bool            `json:"ignoreMissing" yaml:"ignoreMissing"`
	IgnoreValidation *bool            `json:"ignoreValidation" yaml:"ignoreValidation"`
	ImageSource      ImageSource      `json:"imageSource" yaml:"imageSource"`
	IncludeVariables []string         `json:"includeVariables" yaml:"includeVariables"`
	Name             string           `json:"name" yaml:"name"`
	OmitSource       string           `json:"omitSource" yaml:"omitSource"`
	Ports            []Port           `json:"ports" yaml:"ports"`
	Readiness        *HealthCheck     `json:"readiness" yaml:"readiness"`
	Script           string           `json:"script" yaml:"script"`
	SourceIncludes   []string         `json:"sourceIncludes" yaml:"sourceIncludes"`
	SourceExcludes   []string         `json:"sourceExcludes" yaml:"sourceExcludes"`
	SourceLocation   string           `json:"sourceLocation" yaml:"sourceLocation"`
	State            StepState        `json:"state" yaml:"state"`
	Steps            []WorkflowStep   `json:"steps" yaml:"steps"`
	Target           string           `json:"target" yaml:"target"`
	TerminationGrace string           `json:"terminationGrace" yaml:"terminationGrace"`
	Type             string           `json:"type" yaml:"type"`
	Volumes          []Volume         `json:"volumes" yaml:"volumes"`
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
