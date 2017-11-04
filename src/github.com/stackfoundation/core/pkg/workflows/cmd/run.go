package cmd

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/files"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// Run Run a workflow in the current project
func Run(workflowName string) error {
	workflow, err := files.ReadWorkflow(workflowName)
	if err != nil {
		return err
	}

	err = v1.Validate(&workflow.Spec)
	if err != nil {
		return err
	}

	execution, err := execution.NewSyncExecution(workflow)
	if err != nil {
		return err
	}

	execution.Start()
	return nil
}
