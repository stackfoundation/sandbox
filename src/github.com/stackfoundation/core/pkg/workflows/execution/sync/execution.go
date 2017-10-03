package sync

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func handleChangeAndTransitionNext(e execution.Execution, w *v1.Workflow, c *v1.Change) error {
	context := execution.NewContext(w, c)

	log.Debugf("%v event for step %v", c.Type, c.StepSelector)

	switch {
	case context.IsStepReadyToRun():
		return runStepAndTransitionNext(e, context)
	case context.IsGeneratedWorkflowReadyToRun():
		return runGeneratedWorfklowAndTransitionNext(e, context)
	case context.IsWorkflowComplete():
		return e.Complete()
	case context.IsCompoundStepComplete():
		return buildStepImageAndTransitionNext(e, context)
	case context.CanProceedToNextStep():
		return buildStepImageAndTransitionNext(e, context)
	default:
	}

	log.Debugf("Performing no-op for %v event for step %v", c.Type, c.StepSelector)
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
