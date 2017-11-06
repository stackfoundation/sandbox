package controller

import (
	"context"
	"fmt"

	executioncontext "github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/execution/preparation"
	"github.com/stackfoundation/core/pkg/workflows/execution/run"
)

func (l *runListener) Ready(sc *executioncontext.StepContext) {
	l.controller.transitionNext(sc, stepReadyTransition)
}

func (l *runListener) Done(sc *executioncontext.StepContext, r *run.Result) {
	transition := stepDoneTransition{
		variables:          r.Variables,
		generatedContainer: r.Container,
		generatedWorkfow:   r.Workflow,
	}

	l.controller.transitionNext(sc, transition.transition)
}

func (l *runListener) Failed(sc *executioncontext.StepContext, r *run.Result) {
	if !areFailuresIgnored(sc.WorkflowContext.Workflow, sc.Step, sc.StepSelector) {
		fmt.Printf("Step %v failed, aborting!\n", sc.Step.StepName(sc.StepSelector))
		sc.WorkflowContext.Cancel()
		return
	}

	fmt.Printf("Step %v failed, but ignoring and continuing\n", sc.Step.StepName(sc.StepSelector))
	l.Done(sc, r)
}

func (c *executionController) runPodStepAndTransitionNext(sc *executioncontext.StepContext) error {
	err := run.RunPodStep(controller.coordinator, sc, &runListener{controller: c})
	if err != nil {
		if err == context.Canceled {
			return err
		}

		err = shouldIgnoreFailure(sc.WorkflowContext.Workflow, sc.Step, sc.StepSelector, err)
	}

	if err != nil {
		return c.transitionNext(sc, (&stepDoneTransition{}).transition)
	}

	return c.transitionNext(sc, stepStartedTransition)
}

func (c *executionController) runStepAndTransitionNext(sc *executioncontext.StepContext) error {
	err := preparation.PrepareStepIfNecessary(sc.WorkflowContext.Workflow, sc.Step, sc.StepSelector)
	if err != nil {
		return err
	}

	if sc.Step.RequiresBuild() {
		return c.runPodStepAndTransitionNext(sc)
	}

	return c.callExternalWorkflow(sc)
}
