package calls

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/files"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func (controller *Controller) runChildWorkflow(c *Context, workflow *v1.Workflow) error {
	child, err := controller.childExecution(workflow)
	if err != nil {
		err = shouldIgnoreFailure(c.Step, c.StepSelector, c.Workflow, err)
		if err != nil {
			return err
		}

		return controller.transitionNext(c, workflowWaitDoneTransition)
	}

	workflow.Spec.State.Variables = filterVariables(
		c.Step.IncludeVariables,
		c.Step.ExcludeVariables,
		c.Workflow.Spec.State.Variables)

	go func() {
		child.Start()
		log.Debugf("Finished running workflow")
		controller.transitionNext(c, workflowWaitDoneTransition)
	}()

	return controller.transitionNext(c, workflowWaitTransition)
}

func (controller *Controller) runGeneratedWorfklowAndTransitionNext(c *Context) error {
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

	return controller.runChildWorkflow(c, workflow)
}

func (controller *Controller) runExternalWorkflowAndTransitionNext(c *Context) error {
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

	return controller.runChildWorkflow(c, workflow)
}
