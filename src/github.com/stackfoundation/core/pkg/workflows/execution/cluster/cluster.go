package execution

import (
	"context"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type clusterExecution struct {
	workflow   *v1.Workflow
	controller *workflowController
}

// NewClusterExecution Create a new cluster execution for a workflow
func NewClusterExecution(workflow *v1.Workflow) execution.Execution {
	return &clusterExecutor{workflow}
}

func (e *clusterExecutor) Complete() {
	kube.DeleteWorkflow(context.WorkflowsClient(), workflow)
	context.Cancel()
}

func (e *clusterExecutor) Start() {
	clientSet, err := kube.CreateExtensionsClient()
	if err != nil {
		return err
	}

	err = kube.CreateWorkflowResourceDefinitionIfRequired(clientSet.CustomResourceDefinitions())
	if err != nil {
		return err
	}

	err = kube.UploadWorkflow(workflow)
	if err != nil {
		return err
	}

	return RunWorkflowController()
}

func (e *clusterExecutor) UpdateWorkflow(update func(*v1.Workflow)) {
	kube.UpdateWorkflow(context.context.WorkflowsClient(), context.workflow, func(w *v1.Workflow))
}

func (e *clusterExecutor) Workflow() *v1.Workflow {
	return e.workflow
}
