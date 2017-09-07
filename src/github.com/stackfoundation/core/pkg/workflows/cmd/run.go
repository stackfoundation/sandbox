package cmd

import (
	"github.com/stackfoundation/core/pkg/workflows/controller"
	"github.com/stackfoundation/core/pkg/workflows/files"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Run Run a workflow in the current project
func Run(workflowName string) error {
	workflow, err := files.ReadWorkflow(workflowName)
	if err != nil {
		return err
	}

	err := v1.Validate(&workflow.Spec)
	if err != nil {
		return err
	}

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

	return controller.RunWorkflowController()
}
