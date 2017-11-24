package controller

import (
	executioncontext "github.com/stackfoundation/sandbox/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"github.com/stackfoundation/sandbox/log"
)

func (c *executionController) processChangeAndTransitionNext(sc *executioncontext.StepContext) error {
	change := sc.Change

	switch {
	case sc.IsStepReadyToRun():
		return c.runStepAndTransitionNext(sc)
	case sc.IsGeneratedWorkflowReadyToRun():
		return c.callGeneratedWorkflow(sc)
	case sc.IsWorkflowComplete():
		log.Debugf("Workflow completed")
		sc.WorkflowContext.Cancel()
		return nil
	case sc.IsCompoundStepComplete():
		return c.buildStepImageAndTransitionNext(sc)
	case sc.CanProceedToNextStep():
		return c.buildStepImageAndTransitionNext(sc)
	default:
	}

	log.Debugf("Performing no-op for %v event for step %v", change.Type, change.StepSelector)
	return c.transitionNext(sc, consumeTransition)
}

func (c *executionController) processNextChange(wc *executioncontext.WorkflowContext) error {
	if len(wc.Workflow.Spec.State.Changes) < 1 {
		return c.transitionNext(executioncontext.NewStepContext(wc, &v1.Change{}), initialTransition)
	}

	u := wc.Workflow.NextUnhandled()
	if u != nil {
		return c.processChangeAndTransitionNext(executioncontext.NewStepContext(wc, u))
	}

	return nil
}
