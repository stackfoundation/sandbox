package v1

import (
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// ParseWorkflow Parse the given workflow content
func ParseWorkflow(projectRoot, workflowName string, content []byte) (*Workflow, error) {
	var workflowSpec WorkflowSpec
	err := yaml.Unmarshal(content, &workflowSpec)
	if err != nil {
		return nil, err
	}

	workflowSpec.State = WorkflowState{
		ID:          GenerateWorkflowID(),
		ProjectRoot: projectRoot,
		Variables:   CollectVariables(workflowSpec.Variables),
	}

	for i, step := range workflowSpec.Steps {
		if len(step.State.GeneratedImage) > 0 ||
			len(step.State.GeneratedScript) > 0 {
			step.State = StepState{}
			workflowSpec.Steps[i] = step
		}
	}

	workflow := Workflow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: WorkflowsGroupName + "/" + WorkflowsGroupVersion,
			Kind:       WorkflowsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflowName,
			Namespace: v1.NamespaceDefault,
		},
		Spec: workflowSpec,
	}

	return &workflow, nil
}
