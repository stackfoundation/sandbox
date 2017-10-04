package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	log "github.com/stackfoundation/log"
)

// ReadWorkflow Read the workflow with the specified name from the current project directory
func ReadWorkflow(workflowName string) (*workflowsv1.Workflow, error) {
	workflowsDirectory, err := getWorkflowsDirectory()
	if err != nil {
		return nil, err
	}

	workflowFile := filepath.Join(workflowsDirectory, workflowName+workflowExtension)
	log.Debugf(`Looking for workflow "%v" at "%v"`, workflowName, workflowFile)

	workflowFileExists, err := fileExists(workflowFile)
	if err != nil {
		return nil, err
	}

	if !workflowFileExists {
		return nil, os.ErrNotExist
	}

	log.Debugf(`Reading workflow at "%v"`, workflowFile)
	workflowFileContent, err := ioutil.ReadFile(workflowFile)
	if err != nil {
		return nil, err
	}

	projectRoot, err := os.Getwd()

	return workflowsv1.ParseWorkflow(projectRoot, workflowName, workflowFileContent)
}

// DeleteWorkflow Delete the specified workflow from the project
func DeleteWorkflow(workflow string) (bool, error) {
	workflowsDirectory, err := getWorkflowsDirectory()
	if err != nil {
		return false, err
	}

	workflowsDirectoryExists, err := directoryExists(workflowsDirectory)
	if err != nil || !workflowsDirectoryExists {
		return false, err
	}

	workflowFile := filepath.Join(workflowsDirectory, workflow+workflowExtension)

	workflowFileExists, err := fileExists(workflowFile)
	if err != nil || !workflowFileExists {
		return false, err
	}

	err = os.Remove(workflowFile)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ListWorkflows List all workflows in the project
func ListWorkflows() ([]string, error) {
	workflowsDirectory, err := getWorkflowsDirectory()
	if err != nil {
		return []string{}, nil
	}

	workflows, err := ioutil.ReadDir(workflowsDirectory)
	if err != nil {
		return []string{}, nil
	}

	var workflowNames []string
	for _, workflow := range workflows {
		if !workflow.IsDir() && filepath.Ext(workflow.Name()) == workflowExtension {
			workflowNames = append(workflowNames, strings.TrimSuffix(workflow.Name(), workflowExtension))
		}
	}

	return workflowNames, nil
}
