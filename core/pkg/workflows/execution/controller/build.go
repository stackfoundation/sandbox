package controller

import (
	"context"

	executioncontext "github.com/stackfoundation/sandbox/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/image"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/preparation"
)

func (c *executionController) buildStepImageAndTransitionNext(sc *executioncontext.StepContext) error {
	err := preparation.PrepareStepIfNecessary(sc.WorkflowContext.Workflow, sc.NextStep, sc.NextStepSelector)
	if err != nil {
		return err
	}

	if sc.NextStep.RequiresBuild() {
		err := image.BuildStepImage(c.coordinator, sc)
		if err != nil {
			if err == context.Canceled {
				return err
			}

			err = shouldIgnoreFailure(sc.WorkflowContext.Workflow, sc.NextStep, sc.NextStepSelector, err)
			if err != nil {
				return err
			}
		}
	}

	return c.transitionNext(sc, imageBuiltTransition)
}
