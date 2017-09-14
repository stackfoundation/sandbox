package sync

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func shouldProceedToNextStep(c *execution.Context) bool {
	if (c.Change.Type == v1.StepReady && c.Step.IsServiceWithWait()) ||
		(c.Change.Type == v1.StepStarted && (c.Step == nil || c.Step.IsAsync())) ||
		(c.Change.Type == v1.StepDone && !c.Step.IsAsync()) ||
		c.Change.Type == v1.WorkflowWaitDone {
		return true
	}

	return false
}

func handleChangeAndTransitionNext(e execution.Execution, w *v1.Workflow, c *v1.Change) error {
	context := execution.NewContext(w, c)

	if c.Type == v1.StepImageBuilt {
		return runStepAndTransitionNext(e, context)
	}

	if c.Type == v1.StepDone && context.Step.IsGenerator() {
		return runGeneratedWorfklowAndTransitionNext(e, context)
	}

	if context.IsWorkflowComplete() {
		if shouldProceedToNextStep(context) {
			return e.Complete()
		}
	} else {
		if context.IsCompoundStepBoundary() {
			parent := w.Parent(c.StepSelector)
			if parent.IsCompoundStepComplete() {
				return buildStepImageAndTransitionNext(e, context)
			}
		} else {
			if shouldProceedToNextStep(context) {
				return buildStepImageAndTransitionNext(e, context)
			}
		}
	}

	return e.TransitionNext(context, consumeTransition)
}

// Execute Execute by making the next workflow transition
func Execute(e execution.Execution, w *v1.Workflow) error {
	if len(w.Spec.State.Changes) < 1 {
		return e.TransitionNext(&execution.Context{Workflow: w}, initialTransition)
	}

	c := w.NextUnhandled()
	if c != nil {
		return handleChangeAndTransitionNext(e, w, c)
	}

	return nil
}
