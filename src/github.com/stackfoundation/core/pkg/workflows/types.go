package workflows

import (
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/apimachinery/pkg/runtime/schema"
)

type EnvironmentVariable struct {
        Name  string `json:"name"`
        Value string `json:"value"`
}

type WorkflowStep struct {
        StepImage   string `json:"stepImage"`

        Name        string `json:"name"`
        Type        string `json:"type"`
        Image       string `json:"image"`
        ImageSource string `json:"imageSource"`
        Tag         string `json:"tag"`
        Dockerfile  string `json:"dockerfile"`
        Script      string `json:"script"`
        Ports       []string `json:"ports"`
        Variables   []EnvironmentVariable `json:"variables"`
}

type WorkflowStatus struct {
        Step   int `json:"step"`
        Status string `json:"status"`
}

type WorkflowSpec struct {
        ProjectRoot string `json:"projectRoot"`
        File        string `json:"file"`
        Steps       []WorkflowStep `json:"steps"`
        Status      WorkflowStatus `json:"status"`
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
