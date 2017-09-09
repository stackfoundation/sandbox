package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EnvironmentSource Environment source defined for a step
type EnvironmentSource struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
	File  string `json:"file" yaml:"file"`
}

// Volume Volume to mount for a workflow step
type Volume struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	HostPath  string `json:"hostPath" yaml:"hostPath"`
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
	SkipWait bool            `json:"skipWait" yaml:"skipWait"`
	Type     HealthCheckType `json:"type" yaml:"type"`
	Port     int32           `json:"port" yaml:"port"`
	Script   string          `json:"script" yaml:"script"`
	Path     string          `json:"path" yaml:"path"`
	Headers  []HTTPHeader    `json:"headers" yaml:"headers"`
	Interval *int32          `json:"interval" yaml:"interval"`
	Timeout  *int32          `json:"timeout" yaml:"timeout"`
	Retries  *int32          `json:"retries" yaml:"retries"`
	Grace    *int32          `json:"grace" yaml:"grace"`
}

// ImageSource Image source
type ImageSource string

// SourceImage Docker image
const SourceImage ImageSource = "image"

// SourceStep Source is previous step image
const SourceStep ImageSource = "step"

// StepStatus Status of step
type StepStatus string

// StatusStepReady A parallel or service step is ready
const StatusStepReady StepStatus = "stepReady"

// StatusStepDone A parallel or service step is done
const StatusStepDone StepStatus = "stepDone"

// StepState State of step
type StepState struct {
	GeneratedImage  string     `json:"generatedImage" yaml:"generatedImage"`
	GeneratedScript string     `json:"generatedScript" yaml:"generatedScript"`
	Status          StepStatus `json:"status" yaml:"status"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	State            StepState           `json:"state" yaml:"state"`
	Name             string              `json:"name" yaml:"name"`
	Type             string              `json:"type" yaml:"type"`
	OmitSource       bool                `json:"omitSource" yaml:"omitSource"`
	Image            string              `json:"image" yaml:"image"`
	ImageSource      ImageSource         `json:"imageSource" yaml:"imageSource"`
	Dockerfile       string              `json:"dockerfile" yaml:"dockerfile"`
	Script           string              `json:"script" yaml:"script"`
	SourceLocation   string              `json:"sourceLocation" yaml:"sourceLocation"`
	Ports            []string            `json:"ports" yaml:"ports"`
	Readiness        *HealthCheck        `json:"readiness" yaml:"readiness"`
	Health           *HealthCheck        `json:"health" yaml:"health"`
	Environment      []EnvironmentSource `json:"environment" yaml:"environment"`
	Volumes          []Volume            `json:"volumes" yaml:"volumes"`
	Steps            []WorkflowStep      `json:"steps" yaml:"steps"`
	TerminationGrace *int32              `json:"terminationGrace" yaml:"terminationGrace"`
}

// WorkflowStatus Status of workflow
type WorkflowStatus string

// StatusStepImageBuilt Image for step has been built
const StatusStepImageBuilt WorkflowStatus = "imageBuilt"

// StatusCompoundStepFinished A compound step has finished
const StatusCompoundStepFinished WorkflowStatus = "compoundStepFinished"

// StatusStepFinished Running of step has finished
const StatusStepFinished WorkflowStatus = "stepFinished"

// StatusFinished Whole workflows has finished
const StatusFinished WorkflowStatus = "finished"

// WorkflowState State of workflow in K8s
type WorkflowState struct {
	ProjectRoot string         `json:"projectRoot" yaml:"projectRoot"`
	File        string         `json:"file" yaml:"file"`
	Step        []int          `json:"step" yaml:"step"`
	Status      WorkflowStatus `json:"status" yaml:"status"`
}

// WorkflowSpec Specification of workflow
type WorkflowSpec struct {
	State WorkflowState  `json:"state" yaml:"state"`
	Steps []WorkflowStep `json:"steps" yaml:"steps"`
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
const DefaultInterval = 30

// DefaultRetries Default number of retries
const DefaultRetries = 3

// DefaultTimeout Default timeout in seconds
const DefaultTimeout = 30

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
