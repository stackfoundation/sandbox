package controller

import (
	"fmt"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/files"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"github.com/stackfoundation/sandbox/log"
)

func (c *executionController) executeChild(sc *context.StepContext, child *v1.Workflow) error {
	var include []string
	var exclude []string

	source := sc.Step.Source()
	if source != nil {
		include = source.Include
		exclude = source.Exclude
	}

	child.Spec.State.Variables = filterVariables(include, exclude, sc.WorkflowContext.Workflow.Spec.State.Variables)

	go func() {
		c.Execute(sc.WorkflowContext.Context, child)
		log.Debugf("Finished called workflow")
		c.transitionNext(sc, workflowWaitDoneTransition)
	}()

	return c.transitionNext(sc, workflowWaitTransition)
}

func (c *executionController) callGeneratedWorkflow(sc *context.StepContext) error {
	step := sc.Step
	stepName := step.StepName(sc.Change.StepSelector)

	fmt.Println("Running generated workflow:")

	content := []byte(step.State.GeneratedWorkflow)
	workflow, err := v1.ParseWorkflow(sc.WorkflowContext.Workflow.Spec.State.ProjectRoot, stepName, content)
	if err != nil {
		return err
	}

	return c.executeChild(sc, workflow)
}

func (c *executionController) callExternalWorkflow(sc *context.StepContext) error {
	step := sc.Step
	stepName := step.StepName(sc.Change.StepSelector)

	fmt.Println("Running step " + stepName + ":")

	workflow, err := files.ReadWorkflow(sc.Step.External.Workflow)
	if err != nil {
		return err
	}

	return c.executeChild(sc, workflow)
}
