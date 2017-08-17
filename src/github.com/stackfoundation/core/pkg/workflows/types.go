package workflows

import (
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/apimachinery/pkg/runtime/schema"
)

type EnvironmentVariable struct {
        Name  string `json:"name" yaml:"name"`
        Value string `json:"value" yaml:"value"`
}

type WorkflowStep struct {
        StepImage   string `json:"stepImage" yaml:"stepImage"`

        Name        string `json:"name" yaml:"name"`
        Type        string `json:"type" yaml:"type"`
        Image       string `json:"image" yaml:"image"`
        ImageSource string `json:"imageSource" yaml:"imageSource"`
        Tag         string `json:"tag" yaml:"tag"`
        Dockerfile  string `json:"dockerfile" yaml:"dockerfile"`
        Script      string `json:"script" yaml:"script"`
        Ports       []string `json:"ports" yaml:"ports"`
        Variables   []EnvironmentVariable `json:"variables" yaml:"variables"`
}

type WorkflowStatus struct {
        Step   int `json:"step" yaml:"step"`
        Status string `json:"status" yaml:"status"`
}

type WorkflowSpec struct {
        ProjectRoot string `json:"projectRoot" yaml:"projectRoot"`
        File        string `json:"file" yaml:"file"`
        Steps       []WorkflowStep `json:"steps" yaml:"steps"`
        Status      WorkflowStatus `json:"status" yaml:"status"`
}

type Workflow struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata"`
        Spec WorkflowSpec `json:"spec"`
}

type WorkflowList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata"`
        Items []Workflow `json:"items"`
}

const SourceCatalog = "catalog"
const SourceManual = "manual"
const SourceStep = "step"

const StatusStepImageBuilt = "imageBuilt"
const StatusStepFinished = "stepFinished"
const StatusFinished = "finished"

const WorkflowsGroupName = "stack.foundation"
const WorkflowsGroupVersion = "v1"
const WorkflowsKind = "Workflow"
const WorkflowsPluralName = "workflows"

var SchemeGroupVersion = schema.GroupVersion{
        Group: WorkflowsGroupName,
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
