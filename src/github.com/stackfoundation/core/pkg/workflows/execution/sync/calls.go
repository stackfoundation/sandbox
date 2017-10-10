package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/files"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func runChildWorkflow(e execution.Execution, c *execution.Context, workflow *v1.Workflow) error {
	child, err := e.ChildExecution(workflow)
	if err != nil {
		err = shouldIgnoreFailure(c.Step, c.StepSelector, c.Workflow, err)
		if err != nil {
			return err
		}

		return e.TransitionNext(c, workflowWaitDoneTransition)
	}

	workflow.Spec.State.Variables = filterVariables(
		c.Step.IncludeVariables,
		c.Step.ExcludeVariables,
		c.Workflow.Spec.State.Variables)

	go func() {
		child.Start()
		log.Debugf("Finished running workflow")
		e.TransitionNext(c, workflowWaitDoneTransition)
	}()

	return e.TransitionNext(c, workflowWaitTransition)
}

func runGeneratedWorfklowAndTransitionNext(e execution.Execution, c *execution.Context) error {
	step := c.Step
	stepName := step.StepName(c.Change.StepSelector)

	fmt.Println("Running generated workflow:")

	content := []byte(step.State.GeneratedWorkflow)
	workflow, err := v1.ParseWorkflow(c.Workflow.Spec.State.ProjectRoot, stepName, content)
	if err != nil {
		err = shouldIgnoreFailure(c.Step, c.StepSelector, c.Workflow, err)
		if err != nil {
			return err
		}
	}

	return runChildWorkflow(e, c, workflow)
}

func runExternalWorkflowAndTransitionNext(e execution.Execution, c *execution.Context) error {
	step := c.Step
	stepName := step.StepName(c.Change.StepSelector)

	fmt.Println("Running step " + stepName + ":")

	workflow, err := files.ReadWorkflow(c.Step.Target)
	if err != nil {
		err = shouldIgnoreFailure(c.Step, c.StepSelector, c.Workflow, err)
		if err != nil {
			return err
		}
	}

	return runChildWorkflow(e, c, workflow)
}
