package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func runWorkflowStepAndTransitionNext(e execution.Execution, c *execution.Context) error {
	step := c.Step
	stepName := step.StepName(c.Change.StepSelector)

	fmt.Println("Running step " + stepName + ":")

	if len(step.Generator) > 0 {
		content := []byte(step.State.GeneratedWorkflow)
		workflow, err := v1.ParseWorkflow(c.Workflow.Spec.State.ProjectRoot, stepName, content)
		if err != nil {
			return err
		}

		child, err := e.ChildExecution(workflow)
		if err != nil {
			return err
		}

		go child.Start()
	}

	return e.TransitionNext(c, workflowWaitTransition)
}

func runStepAndTransitionNext(e execution.Execution, c *execution.Context) error {
	if c.Step.RequiresBuild() {
		return runPodStepAndTransitionNext(e, c)
	}

	return runWorkflowStepAndTransitionNext(e, c)
}

func (e *syncExecution) RunStep(spec *execution.RunStepSpec) error {
	return kube.CreateAndRunPod(
		e.podsClient,
		&kube.PodCreationSpec{
			LogPrefix:        spec.Name,
			Image:            spec.Image,
			Command:          spec.Command,
			Environment:      spec.Environment,
			Readiness:        spec.Readiness,
			Volumes:          spec.Volumes,
			Context:          e.context,
			Cleanup:          &e.cleanupWaitGroup,
			Listener:         spec.PodListener,
			VariableReceiver: spec.VariableReceiver,
			WorkflowReceiver: spec.WorkflowReceiver,
		})
}
