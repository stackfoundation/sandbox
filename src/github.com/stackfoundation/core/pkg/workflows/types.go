package workflows

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EnvironmentVariable Environment variable defined for a step
type EnvironmentVariable struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// Volume Volume to mount for a workflow step
type Volume struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	HostPath  string `json:"hostPath" yaml:"hostPath"`
}

// WorkflowStep Step within a workflow
type WorkflowStep struct {
	StepImage  string `json:"stepImage" yaml:"stepImage"`
	StepScript string `json:"stepScript" yaml:"stepScript"`

	Name           string                `json:"name" yaml:"name"`
	Type           string                `json:"type" yaml:"type"`
	OmitSource     bool                  `json:"omitSource" yaml:"omitSource"`
	Image          string                `json:"image" yaml:"image"`
	ImageSource    string                `json:"imageSource" yaml:"imageSource"`
	Tag            string                `json:"tag" yaml:"tag"`
	Dockerfile     string                `json:"dockerfile" yaml:"dockerfile"`
	Script         string                `json:"script" yaml:"script"`
	SourceLocation string                `json:"sourceLocation" yaml:"sourceLocation"`
	Ports          []string              `json:"ports" yaml:"ports"`
	Variables      []EnvironmentVariable `json:"variables" yaml:"variables"`
	Volumes        []Volume              `json:"volumes" yaml:"volumes"`
}

// WorkflowStatus Overall status of workflow in K8s
type WorkflowStatus struct {
	Step   int    `json:"step" yaml:"step"`
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

// SourceCatalog Official Docker image
const SourceCatalog = "catalog"

// SourceManual Manual Docker image
const SourceManual = "manual"

// SourceStep Source is previous step image
const SourceStep = "step"

// StatusStepImageBuilt Image for step has been built
const StatusStepImageBuilt = "imageBuilt"

// StatusStepFinished Running of step has finished
const StatusStepFinished = "stepFinished"

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

// SchemeGroupVersion Workflows GroupVersion
var SchemeGroupVersion = schema.GroupVersion{
	Group:   WorkflowsGroupName,
	Version: WorkflowsGroupVersion,
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Workflow{},
		&WorkflowList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
