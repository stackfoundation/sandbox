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

// Health Health checks for a workflow step
type Health struct {
	Type     string       `json:"type" yaml:"type"`
	Port     int32        `json:"port" yaml:"port"`
	Script   string       `json:"script" yaml:"script"`
	Path     string       `json:"path" yaml:"path"`
	Headers  []HTTPHeader `json:"headers" yaml:"headers"`
	Interval *int32       `json:"interval" yaml:"interval"`
	Timeout  *int32       `json:"timeout" yaml:"timeout"`
	Retries  *int32       `json:"retries" yaml:"retries"`
	Grace    *int32       `json:"grace" yaml:"grace"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	StepImage  string `json:"stepImage" yaml:"stepImage"`
	StepScript string `json:"stepScript" yaml:"stepScript"`
	StepStatus string `json:"stepStatus" yaml:"stepStatus"`

	Name           string              `json:"name" yaml:"name"`
	Type           string              `json:"type" yaml:"type"`
	OmitSource     bool                `json:"omitSource" yaml:"omitSource"`
	Image          string              `json:"image" yaml:"image"`
	ImageSource    string              `json:"imageSource" yaml:"imageSource"`
	Dockerfile     string              `json:"dockerfile" yaml:"dockerfile"`
	Script         string              `json:"script" yaml:"script"`
	SourceLocation string              `json:"sourceLocation" yaml:"sourceLocation"`
	Ports          []string            `json:"ports" yaml:"ports"`
	Health         *Health             `json:"health" yaml:"health"`
	Environment    []EnvironmentSource `json:"environment" yaml:"environment"`
	Volumes        []Volume            `json:"volumes" yaml:"volumes"`
	Steps          []WorkflowStep      `json:"steps" yaml:"steps"`
}

// WorkflowStatus Overall status of workflow in K8s
type WorkflowStatus struct {
	Step   []int  `json:"step" yaml:"step"`
	Status string `json:"status" yaml:"status"`
}

// WorkflowSpec Specification of workflow
type WorkflowSpec struct {
	ProjectRoot string         `json:"projectRoot" yaml:"projectRoot"`
	File        string         `json:"file" yaml:"file"`
	Steps       []WorkflowStep `json:"steps" yaml:"steps"`
	Status      WorkflowStatus `json:"status" yaml:"status"`
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

// SourceImage Docker image
const SourceImage = "image"

// SourceStep Source is previous step image
const SourceStep = "step"

// StatusStepImageBuilt Image for step has been built
const StatusStepImageBuilt = "imageBuilt"

// StatusStepFinished Running of step has finished
const StatusStepFinished = "stepFinished"

// StatusStepReady A parallel or service step is ready
const StatusStepReady = "stepReady"

// StatusStepDone A parallel or service step is done
const StatusStepDone = "stepDone"

// StatusFinished Whole workflows has finished
const StatusFinished = "finished"

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

// TCPCheck TCP health check
const TCPCheck = "tcp"

// HTTPCheck HTTP health check
const HTTPCheck = "http"

// HTTPSCheck HTTPS health check
const HTTPSCheck = "https"

// ScriptCheck Script health check
const ScriptCheck = "script"

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
