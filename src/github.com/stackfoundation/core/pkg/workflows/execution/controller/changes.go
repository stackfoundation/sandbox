package controller

import (
	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func (controller *executionController) buildStepImageAndTransitionNext(c *context.ExecutionContext) error {
	err := preparation.PrepareStepIfNecessary(c, c.NextStep, c.NextStepSelector)
	if err != nil {
		return err
	}

	if c.NextStep.RequiresBuild() {
		err := build.BuildStepImage(controller.coordinator, c)
		if err != nil && err != context.Canceled {
			err = shouldIgnoreFailure(c.NextStep, c.NextStepSelector, c.Workflow, err)
			if err != nil {
				return err
			}
		}
	}

	return controller.transitionNext(c, imageBuiltTransition)
}

func (controller *executionController) processChangeAndTransitionNext(w *v1.Workflow, c *v1.Change) error {
	context := context.NewExecutionContext(w, c)

	log.Debugf("%v event for step %v", c.Type, c.StepSelector)

	switch {
	case context.IsStepReadyToRun():
		return controller.runStepAndTransitionNext(context)
	case context.IsGeneratedWorkflowReadyToRun():
		return runGeneratedWorfklowAndTransitionNext(e, context)
	case context.IsWorkflowComplete():
		return e.Complete()
	case context.IsCompoundStepComplete():
		return controller.buildStepImageAndTransitionNext(context)
	case context.CanProceedToNextStep():
		return controller.buildStepImageAndTransitionNext(context)
	default:
	}

	log.Debugf("Performing no-op for %v event for step %v", c.Type, c.StepSelector)
	return controller.transitionNext(context, consumeTransition)
}

func (c *executionController) processNextChange(w *v1.Workflow) error {
	if len(w.Spec.State.Changes) < 1 {
		return c.transitionNext(&Context{Workflow: w}, initialTransition)
	

	u := w.NextUnhandled()
	if u != nil {
		return c.processChangeAndTransitionNext(w, u)
	}

	return nil
}

func (controller *executionController) runPodStepAndTransitionNext(c *context.ExecutionContext) error {
	err := run.RunPodStep(e, c)
	if err != nil {
		return controller.transitionNext(c, (&stepDoneTransition{}).transition)
	}

	return controller.transitionNext(c, stepStartedTransition)
}

func (controller *executionController) runStepAndTransitionNext(c *context.ExecutionContext) error {
	err := preparation.PrepareStepIfNecessary(c, c.Step, c.StepSelector)
	if err != nil {
		return err
	}

	if c.Step.RequiresBuild() {
		return controller.runPodStepAndTransitionNext(c)
	}

	return controller.runExternalWorkflowAndTransitionNext(c)
}
